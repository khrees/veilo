package services_test

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/providers"
)

// Mock Domain Repository
type mockDomainRepository struct {
	domains map[string]*models.Domain
	err     error
}

func (m *mockDomainRepository) Create(d *models.Domain) error {
	if m.err != nil {
		return m.err
	}
	d.ID = uuid.New()
	if m.domains == nil {
		m.domains = make(map[string]*models.Domain)
	}
	m.domains[d.ID.String()] = d
	return nil
}

func (m *mockDomainRepository) Update(d *models.Domain) error {
	if m.err != nil {
		return m.err
	}
	m.domains[d.ID.String()] = d
	return nil
}

func (m *mockDomainRepository) Delete(id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.domains, id)
	return nil
}

func (m *mockDomainRepository) FindAll() ([]models.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]models.Domain, 0, len(m.domains))
	for _, d := range m.domains {
		result = append(result, *d)
	}
	return result, nil
}

func (m *mockDomainRepository) FindByID(id string) (*models.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}
	d, exists := m.domains[id]
	if !exists {
		return nil, errors.New("domain not found")
	}
	return d, nil
}

func (m *mockDomainRepository) FindByName(name string) (*models.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, d := range m.domains {
		if d.Name == name {
			return d, nil
		}
	}
	return nil, errors.New("domain not found")
}

// Mock Alias Repository
type mockAliasRepository struct {
	aliases map[string]*models.Alias
	err     error
}

func (m *mockAliasRepository) Create(a *models.Alias) error {
	if m.err != nil {
		return m.err
	}
	a.ID = uuid.New()
	if m.aliases == nil {
		m.aliases = make(map[string]*models.Alias)
	}
	m.aliases[a.ID.String()] = a
	return nil
}

func (m *mockAliasRepository) FindAll(filter models.AliasFilter) ([]models.Alias, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]models.Alias, 0, len(m.aliases))
	for _, a := range m.aliases {
		if filter.Enabled != nil && a.Enabled != *filter.Enabled {
			continue
		}
		if filter.Domain != nil && *filter.Domain != "" && a.Domain != *filter.Domain {
			continue
		}
		result = append(result, *a)
	}
	if filter.Offset != nil && *filter.Offset >= 0 && *filter.Offset < len(result) {
		result = result[*filter.Offset:]
	} else if filter.Offset != nil && *filter.Offset >= len(result) {
		result = []models.Alias{}
	}
	if filter.Limit != nil && *filter.Limit >= 0 && *filter.Limit < len(result) {
		result = result[:*filter.Limit]
	}
	return result, nil
}

func (m *mockAliasRepository) FindByID(id string) (*models.Alias, error) {
	if m.err != nil {
		return nil, m.err
	}
	a, exists := m.aliases[id]
	if !exists {
		return nil, errors.New("alias not found")
	}
	return a, nil
}

func (m *mockAliasRepository) Update(id string, updates map[string]any) error {
	if m.err != nil {
		return m.err
	}
	alias, exists := m.aliases[id]
	if !exists {
		return errors.New("alias not found")
	}
	for key, value := range updates {
		switch key {
		case "real_email":
			alias.RealEmail = value.(string)
		case "enabled":
			alias.Enabled = value.(bool)
		}
	}
	return nil
}

func (m *mockAliasRepository) Delete(id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.aliases, id)
	return nil
}

func (m *mockAliasRepository) FindByAddress(address string) (*models.Alias, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, a := range m.aliases {
		if a.Address == address {
			return a, nil
		}
	}
	return nil, errors.New("alias not found")
}

func (m *mockAliasRepository) DisableExpired(now time.Time) error {
	if m.err != nil {
		return m.err
	}
	for _, a := range m.aliases {
		if a.ExpiresAt != nil && a.ExpiresAt.Before(now) {
			a.Enabled = false
		}
	}
	return nil
}

// Mock ForwardLog Repository
type mockForwardLogRepository struct {
	logs  map[string]*models.ForwardLog
	stats *models.Stats
	err   error
}

func (m *mockForwardLogRepository) FindByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []models.ForwardLog
	for _, log := range m.logs {
		if log.AliasID.String() == aliasID {
			result = append(result, *log)
		}
	}
	return result, nil
}

func (m *mockForwardLogRepository) GetStats() (*models.Stats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

func (m *mockForwardLogRepository) Create(f *models.ForwardLog) error {
	if m.err != nil {
		return m.err
	}
	f.ID = uuid.New()
	if m.logs == nil {
		m.logs = make(map[string]*models.ForwardLog)
	}
	m.logs[f.ID.String()] = f
	return nil
}

// Mock Email Provider
type mockEmailProvider struct {
	registerDomainFunc   func(ctx context.Context, domainName string) (*providers.RegisterDomainResult, error)
	verifyDomainFunc     func(ctx context.Context, domainID string) (bool, error)
	getReceivedEmailFunc func(ctx context.Context, emailID string) (*providers.ReceivedEmail, error)
	sendEmailFunc        func(ctx context.Context, input providers.SendEmailInput) (string, error)
	ensureWebhookFunc    func(ctx context.Context, webhookURL string) (string, string, error)
}

func (m *mockEmailProvider) RegisterDomain(ctx context.Context, domainName string) (*providers.RegisterDomainResult, error) {
	if m.registerDomainFunc != nil {
		return m.registerDomainFunc(ctx, domainName)
	}
	return nil, nil
}

func (m *mockEmailProvider) VerifyDomain(ctx context.Context, domainID string) (bool, error) {
	if m.verifyDomainFunc != nil {
		return m.verifyDomainFunc(ctx, domainID)
	}
	return false, nil
}

func (m *mockEmailProvider) GetReceivedEmail(ctx context.Context, emailID string) (*providers.ReceivedEmail, error) {
	if m.getReceivedEmailFunc != nil {
		return m.getReceivedEmailFunc(ctx, emailID)
	}
	return nil, nil
}

func (m *mockEmailProvider) SendEmail(ctx context.Context, input providers.SendEmailInput) (string, error) {
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(ctx, input)
	}
	return "", nil
}

func (m *mockEmailProvider) EnsureWebhook(ctx context.Context, webhookURL string) (string, string, error) {
	if m.ensureWebhookFunc != nil {
		return m.ensureWebhookFunc(ctx, webhookURL)
	}
	return "", "", nil
}

// Mock DNS Provider
type mockDNSProvider struct {
	configureDNSFunc func(ctx context.Context, domainName string, records []providers.DNSRecord) error
}

func (m *mockDNSProvider) ConfigureDNS(ctx context.Context, domainName string, records []providers.DNSRecord) error {
	if m.configureDNSFunc != nil {
		return m.configureDNSFunc(ctx, domainName, records)
	}
	return nil
}
