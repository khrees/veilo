package providers

import (
	"context"
	"strconv"

	"github.com/resend/resend-go/v3"
)

type resendEmailProvider struct {
	client *resend.Client
}

// NewResendEmailProvider creates a new EmailProvider backed by Resend
func NewResendEmailProvider(client *resend.Client) EmailProvider {
	return &resendEmailProvider{
		client: client,
	}
}

func (p *resendEmailProvider) RegisterDomain(ctx context.Context, domainName string) (*RegisterDomainResult, error) {
	res, err := p.client.Domains.CreateWithContext(ctx, &resend.CreateDomainRequest{
		Name: domainName,
	})
	if err != nil {
		return nil, err
	}

	records := make([]EmailRecord, 0, len(res.Records))
	for _, rec := range res.Records {
		var priority int
		if rec.Priority.String() != "" {
			pr, err := strconv.Atoi(rec.Priority.String())
			if err == nil {
				priority = pr
			}
		}
		records = append(records, EmailRecord{
			Type:     rec.Type,
			Name:     rec.Name,
			Value:    rec.Value,
			Priority: priority,
		})
	}

	return &RegisterDomainResult{
		DomainID: res.Id,
		Records:  records,
		Verified: res.Status == "verified",
	}, nil
}

func (p *resendEmailProvider) VerifyDomain(ctx context.Context, domainID string) (bool, error) {
	_, err := p.client.Domains.VerifyWithContext(ctx, domainID)
	if err != nil {
		return false, err
	}

	res, err := p.client.Domains.GetWithContext(ctx, domainID)
	if err != nil {
		return false, err
	}

	return res.Status == "verified", nil
}

func (p *resendEmailProvider) GetReceivedEmail(ctx context.Context, emailID string) (*ReceivedEmail, error) {
	res, err := p.client.Emails.Receiving.GetWithContext(ctx, emailID)
	if err != nil {
		return nil, err
	}

	return &ReceivedEmail{
		ID:      res.Id,
		Subject: res.Subject,
		Html:    res.Html,
		Text:    res.Text,
	}, nil
}

func (p *resendEmailProvider) SendEmail(ctx context.Context, input SendEmailInput) (string, error) {
	params := &resend.SendEmailRequest{
		From:    input.From,
		To:      input.To,
		Subject: input.Subject,
		Html:    input.Html,
		Text:    input.Text,
		Headers: input.Headers,
	}

	res, err := p.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return "", err
	}

	return res.Id, nil
}
