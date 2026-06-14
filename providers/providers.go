// Package providers defines interfaces and implementations for external service providers (e.g. DNS, Email).
package providers

import "context"

type EmailRecord struct {
	Type     string
	Name     string
	Value    string
	Priority int
}

type RegisterDomainResult struct {
	DomainID string
	Records  []EmailRecord
	Verified bool
}

type SendEmailInput struct {
	From    string
	To      []string
	Subject string
	Html    string
	Text    string
	Headers map[string]string
}

type ReceivedEmail struct {
	ID      string
	Subject string
	Html    string
	Text    string
}

type EmailProvider interface {
	RegisterDomain(ctx context.Context, domainName string) (*RegisterDomainResult, error)
	VerifyDomain(ctx context.Context, domainID string) (bool, error)
	GetReceivedEmail(ctx context.Context, emailID string) (*ReceivedEmail, error)
	SendEmail(ctx context.Context, input SendEmailInput) (string, error)
	EnsureWebhook(ctx context.Context, webhookURL string) (string, string, error)
}

type DNSRecord struct {
	Type     string
	Name     string
	Value    string
	Priority int
}

type DNSProvider interface {
	ConfigureDNS(ctx context.Context, domainName string, records []DNSRecord) error
}
