// Package services contains the service layer for the application.
package services

import (
	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
)

// --------------------
// Domain Service
// --------------------

// IDomainService interface for domain operations
type IDomainService interface {
	Register(domainName string) error
	Remove(domainName string) error
	FindAll() ([]models.Domain, error)
	FindByID(id string) (*models.Domain, error)
	FindByDomain(domain string) (*models.Domain, error)
}

// domainService implements IDomainService
type domainService struct {
	domainRepo repositories.DomainRepository
}

// NewDomainService will instantiate DomainService
func NewDomainService(domainRepo repositories.DomainRepository) IDomainService {
	return &domainService{
		domainRepo: domainRepo,
	}
}

// Register creates a new domain
func (d *domainService) Register(domainName string) error {
	domain := &models.Domain{
		Domain: domainName,
	}

	return d.domainRepo.Create(domain)
}

// Remove deletes a domain by name
func (d *domainService) Remove(domainName string) error {
	domain, err := d.domainRepo.FindByDomain(domainName)
	if err != nil {
		return err
	}

	return d.domainRepo.Delete(domain.ID.String())
}

// FindAll returns all domains
func (d *domainService) FindAll() ([]models.Domain, error) {
	return d.domainRepo.FindAll()
}

// FindByID finds a domain by ID
func (d *domainService) FindByID(id string) (*models.Domain, error) {
	return d.domainRepo.FindByID(id)
}

// FindByDomain finds a domain by domain name
func (d *domainService) FindByDomain(domain string) (*models.Domain, error) {
	return d.domainRepo.FindByDomain(domain)
}

// --------------------
// Alias Service
// --------------------

// IAliasService interface for alias operations
type IAliasService interface {
	Create(input AliasCreateInput) (*models.Alias, error)
	GetAll() ([]models.Alias, error)
	GetByID(id string) (*models.Alias, error)
	Update(id string, updates map[string]any) error
	Delete(id string) error
}

// AliasCreateInput groups the values needed to create an alias.
type AliasCreateInput struct {
	Address   string
	Slug      string
	Domain    string
	RealEmail string
	Label     *string
	Enabled   bool
}

// aliasService implements IAliasService
type aliasService struct {
	aliasRepo repositories.AliasRepository
}

// NewAliasService will instantiate AliasService
func NewAliasService(aliasRepo repositories.AliasRepository) IAliasService {
	return &aliasService{
		aliasRepo: aliasRepo,
	}
}

// Create creates a new alias
func (a *aliasService) Create(input AliasCreateInput) (*models.Alias, error) {
	alias := &models.Alias{
		Address:      input.Address,
		Slug:         input.Slug,
		Domain:       input.Domain,
		RealEmail:    input.RealEmail,
		Label:        input.Label,
		Enabled:      input.Enabled,
		ForwardCount: 0,
	}

	err := a.aliasRepo.Create(alias)
	if err != nil {
		return nil, err
	}

	return alias, nil
}

// GetAll returns all aliases
func (a *aliasService) GetAll() ([]models.Alias, error) {
	return a.aliasRepo.FindAll()
}

// GetByID returns an alias by ID
func (a *aliasService) GetByID(id string) (*models.Alias, error) {
	return a.aliasRepo.FindByID(id)
}

// Update modifies an existing alias
func (a *aliasService) Update(id string, updates map[string]any) error {
	return a.aliasRepo.Update(id, updates)
}

// Delete removes an alias
func (a *aliasService) Delete(id string) error {
	return a.aliasRepo.Delete(id)
}

// --------------------
// ForwardLog Service
// --------------------

// IForwardLogService interface for forward log operations
type IForwardLogService interface {
	GetByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error)
	GetStats() (*repositories.Stats, error)
}

// forwardLogService implements IForwardLogService
type forwardLogService struct {
	forwardLogRepo repositories.ForwardLogRepository
}

// NewForwardLogService will instantiate ForwardLogService
func NewForwardLogService(forwardLogRepo repositories.ForwardLogRepository) IForwardLogService {
	return &forwardLogService{
		forwardLogRepo: forwardLogRepo,
	}
}

// GetByAliasID returns forward logs for a specific alias
func (f *forwardLogService) GetByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error) {
	return f.forwardLogRepo.FindByAliasID(aliasID, limit, offset)
}

// GetStats returns statistics
func (f *forwardLogService) GetStats() (*repositories.Stats, error) {
	return f.forwardLogRepo.GetStats()
}

// --------------------
// ReplyToken Service
// --------------------

// IReplyTokenService interface for reply token operations
type IReplyTokenService interface {
	Create(input ReplyTokenCreateInput) (*models.ReplyToken, error)
	Find(token string) (*models.ReplyToken, error)
	Delete(token string) error
}

// ReplyTokenCreateInput groups the values needed to create a reply token.
type ReplyTokenCreateInput struct {
	AliasID         uuid.UUID
	OriginalSender  string
	OriginalSubject string
	ThreadID        string
	ExpiresAt       int
}

// replyTokenService implements IReplyTokenService
type replyTokenService struct {
	// replyTokenRepo repositories.ReplyTokenRepository
}

// NewReplyTokenService will instantiate ReplyTokenService
func NewReplyTokenService() IReplyTokenService {
	return &replyTokenService{}
}

// Create creates a new reply token
func (r *replyTokenService) Create(input ReplyTokenCreateInput) (*models.ReplyToken, error) {
	// TODO: Implement token creation with expiration
	return nil, nil
}

// Find finds a reply token
func (r *replyTokenService) Find(token string) (*models.ReplyToken, error) {
	// TODO: Implement token lookup
	return nil, nil
}

// Delete deletes a reply token
func (r *replyTokenService) Delete(token string) error {
	// TODO: Implement token deletion
	return nil
}
