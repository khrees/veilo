package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/providers"
	"github.com/khrees/veilo/services"
	"github.com/stretchr/testify/assert"
)

func TestDomainService_Register(t *testing.T) {
	t.Parallel()
	mockRepo := &mockDomainRepository{}
	svc := services.NewDomainService(mockRepo, nil, nil)

	err := svc.Register("test.com")
	assert.NoError(t, err)

	domains, err := mockRepo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, domains, 1)
	assert.Equal(t, "test.com", domains[0].Name)
}

func TestDomainService_Remove(t *testing.T) {
	t.Parallel()
	mockRepo := &mockDomainRepository{}
	svc := services.NewDomainService(mockRepo, nil, nil)

	err := svc.Register("test.com")
	assert.NoError(t, err)

	err = svc.Remove("test.com")
	assert.NoError(t, err)

	domains, err := mockRepo.FindAll()
	assert.NoError(t, err)
	assert.Empty(t, domains)
}

func TestDomainService_FindAll(t *testing.T) {
	t.Parallel()
	mockRepo := &mockDomainRepository{
		domains: map[string]*models.Domain{
			"1": {ID: uuid.New(), Name: "domain1.com", Verified: true},
			"2": {ID: uuid.New(), Name: "domain2.com", Verified: false},
		},
	}
	svc := services.NewDomainService(mockRepo, nil, nil)

	domains, err := svc.FindAll()
	assert.NoError(t, err)
	assert.Len(t, domains, 2)
}

func TestDomainService_FindByID(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	mockRepo := &mockDomainRepository{
		domains: map[string]*models.Domain{
			id.String(): {ID: id, Name: "test.com", Verified: true},
		},
	}
	svc := services.NewDomainService(mockRepo, nil, nil)

	domain, err := svc.FindByID(id.String())
	assert.NoError(t, err)
	assert.Equal(t, "test.com", domain.Name)
}

func TestDomainService_FindByName(t *testing.T) {
	t.Parallel()
	mockRepo := &mockDomainRepository{
		domains: map[string]*models.Domain{
			"1": {Name: "test.com", Verified: true},
		},
	}
	svc := services.NewDomainService(mockRepo, nil, nil)

	domain, err := svc.FindByName("test.com")
	assert.NoError(t, err)
	assert.Equal(t, "test.com", domain.Name)
}

func TestDomainService_VerifyDomains(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	mockRepo := &mockDomainRepository{
		domains: map[string]*models.Domain{
			id.String(): {ID: id, Name: "test.com", Verified: false},
		},
	}

	mockEmail := &mockEmailProvider{
		registerDomainFunc: func(ctx context.Context, name string) (*providers.RegisterDomainResult, error) {
			return &providers.RegisterDomainResult{
				DomainID: "dom_123",
				Verified: false,
			}, nil
		},
		verifyDomainFunc: func(ctx context.Context, domID string) (bool, error) {
			return true, nil
		},
	}

	svc := services.NewDomainService(mockRepo, mockEmail, nil)
	err := svc.VerifyDomains(context.Background())
	assert.NoError(t, err)

	domain, err := mockRepo.FindByID(id.String())
	assert.NoError(t, err)
	assert.True(t, domain.Verified)
}

func TestDomainService_ErrorHandling(t *testing.T) {
	t.Parallel()
	errTest := errors.New("repository error")

	mockRepo := &mockDomainRepository{err: errTest}
	svc := services.NewDomainService(mockRepo, nil, nil)

	err := svc.Register("test.com")
	assert.ErrorIs(t, err, errTest)

	_, err = svc.FindAll()
	assert.ErrorIs(t, err, errTest)
}

func TestDomainService_Register_WithProviders(t *testing.T) {
	t.Parallel()
	mockRepo := &mockDomainRepository{}

	mockEmail := &mockEmailProvider{
		registerDomainFunc: func(ctx context.Context, name string) (*providers.RegisterDomainResult, error) {
			return &providers.RegisterDomainResult{
				DomainID: "dom_123",
				Records: []providers.EmailRecord{
					{Type: "MX", Name: "@", Value: "feedback-smtp.us-east-1.amazonses.com", Priority: 10},
				},
				Verified: false,
			}, nil
		},
		verifyDomainFunc: func(ctx context.Context, domID string) (bool, error) {
			return true, nil
		},
	}

	dnsCalled := false
	mockDNS := &mockDNSProvider{
		configureDNSFunc: func(ctx context.Context, domainName string, records []providers.DNSRecord) error {
			dnsCalled = true
			assert.Len(t, records, 1)
			return nil
		},
	}

	svc := services.NewDomainService(mockRepo, mockEmail, mockDNS)
	err := svc.Register("test.com")
	assert.NoError(t, err)
	assert.True(t, dnsCalled)

	domains, err := mockRepo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, domains, 1)
	assert.True(t, domains[0].Verified)
}
