package services

import "github.com/khrees/veilo/models"
import "github.com/khrees/veilo/repositories"

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
