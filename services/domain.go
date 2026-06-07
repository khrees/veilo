// Package services contains the service layer for the application.
package services

import (
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
)

// IDomainService interface for domain operations
type IDomainService interface {
	Register(domainName string) error
	Remove(domainName string) error
	FindAll() ([]models.Domain, error)
	FindByID(id string) (*models.Domain, error)
	FindByName(name string) (*models.Domain, error)
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
		Name: domainName,
	}

	return d.domainRepo.Create(domain)
}

// Remove deletes a domain by name
func (d *domainService) Remove(domainName string) error {
	domain, err := d.domainRepo.FindByName(domainName)
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

// FindByName finds a domain by name
func (d *domainService) FindByName(name string) (*models.Domain, error) {
	return d.domainRepo.FindByName(name)
}
