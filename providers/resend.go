package providers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	var domainID string

	// 1. Try to create the domain with receiving capability
	res, err := p.client.Domains.CreateWithContext(ctx, &resend.CreateDomainRequest{
		Name: domainName,
		Capabilities: &resend.DomainCapabilities{
			Sending:   "enabled",
			Receiving: "enabled",
		},
	})
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(strings.ToLower(errMsg), "already") || strings.Contains(strings.ToLower(errMsg), "registered") {
			// Fallback: list domains to find the ID of the existing domain
			list, listErr := p.client.Domains.ListWithContext(ctx)
			if listErr != nil {
				return nil, err
			}
			var found bool
			for _, dom := range list.Data {
				if dom.Name == domainName {
					domainID = dom.Id
					found = true
					break
				}
			}
			if !found {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		domainID = res.Id
	}

	// 2. Always ensure receiving capability is enabled (PATCH API fallback)
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey != "" {
		patchURL := fmt.Sprintf("https://api.resend.com/domains/%s", domainID)
		body := []byte(`{"capabilities": {"receiving": "enabled"}}`)
		req, patchErr := http.NewRequestWithContext(ctx, "PATCH", patchURL, bytes.NewReader(body))
		if patchErr == nil {
			req.Header.Set("Authorization", "Bearer "+apiKey)
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 10 * time.Second}
			resp, doErr := client.Do(req)
			if doErr == nil {
				resp.Body.Close()
			}
		}
	}

	// 3. Fetch the latest domain details (so we get the newly generated inbound MX record)
	fullDom, err := p.client.Domains.GetWithContext(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch domain details from Resend: %w", err)
	}

	records := make([]EmailRecord, 0, len(fullDom.Records))
	for _, rec := range fullDom.Records {
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
		DomainID: domainID,
		Records:  records,
		Verified: fullDom.Status == "verified",
	}, nil
}

func (p *resendEmailProvider) VerifyDomain(ctx context.Context, domainID string) (bool, error) {
	res, err := p.client.Domains.GetWithContext(ctx, domainID)
	if err != nil {
		return false, err
	}

	if res.Status == "verified" {
		return true, nil
	}

	_, err = p.client.Domains.VerifyWithContext(ctx, domainID)
	if err != nil {
		return false, err
	}

	res, err = p.client.Domains.GetWithContext(ctx, domainID)
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

func (p *resendEmailProvider) EnsureWebhook(ctx context.Context, webhookURL string) (string, string, error) {
	list, err := p.client.Webhooks.ListWithContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to list webhooks: %w", err)
	}

	for _, wh := range list.Data {
		if wh.Endpoint == webhookURL {
			return wh.Id, "", nil
		}
	}

	params := &resend.CreateWebhookRequest{
		Endpoint: webhookURL,
		Events:   []string{"email.received", "email.bounced"},
	}

	newWh, err := p.client.Webhooks.CreateWithContext(ctx, params)
	if err != nil {
		return "", "", fmt.Errorf("failed to create webhook: %w", err)
	}

	return newWh.Id, newWh.SigningSecret, nil
}
