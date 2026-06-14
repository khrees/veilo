package services

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/providers"
	"github.com/khrees/veilo/repositories"
)

type EmailReceivedInput struct {
	EmailID   string
	From      string
	To        []string
	MessageID string
	Subject   string
}

type IWebhookService interface {
	ProcessEmailReceived(ctx context.Context, input EmailReceivedInput) error
}

type webhookService struct {
	aliasRepo         repositories.AliasRepository
	forwardLogRepo    repositories.ForwardLogRepository
	replyTokenRepo    repositories.ReplyTokenRepository
	emailProv         providers.EmailProvider
	replyTokenTTLDays int
}

func NewWebhookService(
	aliasRepo repositories.AliasRepository,
	forwardLogRepo repositories.ForwardLogRepository,
	replyTokenRepo repositories.ReplyTokenRepository,
	emailProv providers.EmailProvider,
	replyTokenTTLDays int,
) IWebhookService {
	return &webhookService{
		aliasRepo:         aliasRepo,
		forwardLogRepo:    forwardLogRepo,
		replyTokenRepo:    replyTokenRepo,
		emailProv:         emailProv,
		replyTokenTTLDays: replyTokenTTLDays,
	}
}

// replyRegex matches reply+<token>@domain
var replyRegex = regexp.MustCompile(`^reply\+([^@]+)@(.+)$`)

func (s *webhookService) ProcessEmailReceived(ctx context.Context, input EmailReceivedInput) error {
	if len(input.To) == 0 {
		return fmt.Errorf("no 'to' addresses in email payload")
	}

	toAddress := extractEmailAddress(input.To[0])
	token, isReply := extractReplyToken(toAddress)

	if isReply {
		return s.handleReplyFlow(ctx, input, token)
	}

	return s.handleForwardFlow(ctx, input, toAddress)
}

func extractEmailAddress(raw string) string {
	parsed, err := mail.ParseAddress(raw)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(raw))
	}
	return strings.ToLower(strings.TrimSpace(parsed.Address))
}

func extractReplyToken(address string) (string, bool) {
	matches := replyRegex.FindStringSubmatch(address)
	if len(matches) == 3 {
		return matches[1], true
	}
	return "", false
}

func (s *webhookService) handleForwardFlow(ctx context.Context, input EmailReceivedInput, toAddress string) error {
	// 4. Lookup alias by address in Postgres
	alias, err := s.aliasRepo.FindByAddress(toAddress)

	// 5. Not found or disabled → log as blocked, return nil (silent drop)
	if err != nil || alias == nil || !alias.Enabled {
		log.Warnf("Forward blocked: alias %s not found or disabled (err: %v)", toAddress, err)

		if alias != nil && !alias.Enabled {
			senderEmail := extractEmailAddress(input.From)
			subject := input.Subject
			logEntry := &models.ForwardLog{
				AliasID:   alias.ID,
				Direction: "inbound",
				Sender:    &senderEmail,
				Subject:   &subject,
				Status:    "blocked",
			}
			if logErr := s.forwardLogRepo.Create(logEntry); logErr != nil {
				log.Errorf("failed to create blocked forward log: %v", logErr)
			}
		}

		return nil
	}

	// 6. Rewrite headers:
	parsedFrom, err := mail.ParseAddress(input.From)
	senderName := ""
	senderEmail := input.From
	if err == nil {
		senderName = parsedFrom.Name
		senderEmail = parsedFrom.Address
	}
	if senderName == "" {
		parts := strings.Split(senderEmail, "@")
		if len(parts) > 0 {
			senderName = parts[0]
		} else {
			senderName = senderEmail
		}
	}
	senderName = strings.ReplaceAll(senderName, "\"", "")

	// 7. Create reply_token row in Postgres
	expiresAt := time.Now().Add(time.Duration(s.replyTokenTTLDays) * 24 * time.Hour)
	token := uuid.NewString()

	var subjectPtr *string
	if input.Subject != "" {
		subjectPtr = &input.Subject
	}
	var threadIDPtr *string
	if input.MessageID != "" {
		threadIDPtr = &input.MessageID
	}

	replyToken := &models.ReplyToken{
		Token:           token,
		AliasID:         alias.ID,
		OriginalSender:  senderEmail,
		OriginalSubject: subjectPtr,
		ThreadID:        threadIDPtr,
		ExpiresAt:       expiresAt,
	}

	if createErr := s.replyTokenRepo.Create(replyToken); createErr != nil {
		log.Errorf("failed to create reply token: %v", createErr)
		return fmt.Errorf("failed to process reply token: %w", createErr)
	}

	newFrom := fmt.Sprintf("\"%s via Cloakee\" <%s>", senderName, alias.Address)
	replyTo := fmt.Sprintf("reply+%s@%s", replyToken.Token, alias.Domain)

	headersMap := map[string]string{
		"Reply-To":        replyTo,
		"X-Original-From": senderEmail,
		"X-Forwarded-To":  alias.Address,
	}

	// Fetch full received email content via generic email provider
	receivedEmail, err := s.emailProv.GetReceivedEmail(ctx, input.EmailID)
	if err != nil {
		log.Errorf("failed to fetch received email from provider (ID: %s): %v", input.EmailID, err)
		return fmt.Errorf("failed to fetch email content: %w", err)
	}

	// 8. Send via email provider to alias.real_email
	_, err = s.emailProv.SendEmail(ctx, providers.SendEmailInput{
		From:    newFrom,
		To:      []string{alias.RealEmail},
		Subject: receivedEmail.Subject,
		Html:    receivedEmail.Html,
		Text:    receivedEmail.Text,
		Headers: headersMap,
	})
	if err != nil {
		log.Errorf("failed to forward email via provider: %v", err)
		return fmt.Errorf("failed to forward email: %w", err)
	}

	// 9. Update aliases.forward_count, aliases.last_used_at
	updates := map[string]any{
		"forward_count": alias.ForwardCount + 1,
		"last_used_at":  time.Now(),
	}
	if updateErr := s.aliasRepo.Update(alias.ID.String(), updates); updateErr != nil {
		log.Errorf("failed to update alias stats: %v", updateErr)
	}

	// 10. Insert forward_log row (direction=inbound, status=delivered)
	logEntry := &models.ForwardLog{
		AliasID:   alias.ID,
		Direction: "inbound",
		Sender:    &senderEmail,
		Subject:   &receivedEmail.Subject,
		Status:    "delivered",
	}
	if logErr := s.forwardLogRepo.Create(logEntry); logErr != nil {
		log.Errorf("failed to create forward log: %v", logErr)
	}

	return nil
}

func (s *webhookService) handleReplyFlow(ctx context.Context, input EmailReceivedInput, token string) error {
	// 5. Lookup reply_token in Postgres
	replyToken, err := s.replyTokenRepo.FindByToken(token)

	// 6. Not found or expired → drop silently
	if err != nil || replyToken == nil || replyToken.ExpiresAt.Before(time.Now()) {
		log.Warnf("Reply blocked: token %s not found or expired (err: %v)", token, err)
		return nil
	}

	// Fetch alias
	alias, err := s.aliasRepo.FindByID(replyToken.AliasID.String())
	if err != nil || alias == nil || !alias.Enabled {
		log.Warnf("Reply blocked: alias not found or disabled for token %s (err: %v)", token, err)
		return nil
	}

	// Fetch full received email content via generic email provider
	receivedEmail, err := s.emailProv.GetReceivedEmail(ctx, input.EmailID)
	if err != nil {
		log.Errorf("failed to fetch received email from provider (ID: %s): %v", input.EmailID, err)
		return fmt.Errorf("failed to fetch email content: %w", err)
	}

	// 7. Rewrite headers:
	//    - From: alias address
	//    - To:   reply_token.original_sender
	// 8. Send via email provider
	_, err = s.emailProv.SendEmail(ctx, providers.SendEmailInput{
		From:    alias.Address,
		To:      []string{replyToken.OriginalSender},
		Subject: receivedEmail.Subject,
		Html:    receivedEmail.Html,
		Text:    receivedEmail.Text,
	})
	if err != nil {
		log.Errorf("failed to send reply email via provider: %v", err)
		return fmt.Errorf("failed to send reply: %w", err)
	}

	// 9. Insert forward_log row (direction=reply, status=delivered)
	senderEmail := extractEmailAddress(input.From)
	logEntry := &models.ForwardLog{
		AliasID:   alias.ID,
		Direction: "reply",
		Sender:    &senderEmail,
		Subject:   &receivedEmail.Subject,
		Status:    "delivered",
	}
	if logErr := s.forwardLogRepo.Create(logEntry); logErr != nil {
		log.Errorf("failed to create forward log: %v", logErr)
	}

	return nil
}
