// Package services contains the service layer for the application.
package services

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/providers"
	"github.com/khrees/veilo/repositories"
)

// DomainService interface for domain operations
type DomainService interface {
	Register(domainName string) error
	Remove(domainName string) error
	FindAll() ([]models.Domain, error)
	FindByID(id string) (*models.Domain, error)
	FindByName(name string) (*models.Domain, error)
	VerifyDomains(ctx context.Context) error
}

// domainService implements DomainService
type domainService struct {
	domainRepo repositories.DomainRepository
	emailProv  providers.EmailProvider
	dnsProv    providers.DNSProvider
}

// NewDomainService will instantiate DomainService using generic providers
func NewDomainService(
	domainRepo repositories.DomainRepository,
	emailProv providers.EmailProvider,
	dnsProv providers.DNSProvider,
) DomainService {
	return &domainService{
		domainRepo: domainRepo,
		emailProv:  emailProv,
		dnsProv:    dnsProv,
	}
}

// Register creates a new domain, registers in email provider, configures DNS, and triggers verification
func (d *domainService) Register(domainName string) error {
	// If emailProv is nil, fallback to database-only registration
	if d.emailProv == nil {
		domain := &models.Domain{
			Name:     domainName,
			Verified: false,
		}
		log.Info("please configure an email provider, falling back to database only creation")
		return d.domainRepo.Create(domain)
	}

	ctx := context.Background()

	// 1. Register domain in email provider
	res, err := d.emailProv.RegisterDomain(ctx, domainName)
	if err != nil {
		return fmt.Errorf("failed to register domain in email provider: %w", err)
	}

	// 2. Configure DNS if dnsProv is set
	if d.dnsProv != nil {
		dnsRecords := make([]providers.DNSRecord, 0, len(res.Records))
		for _, rec := range res.Records {
			dnsRecords = append(dnsRecords, providers.DNSRecord{
				Type:     rec.Type,
				Name:     rec.Name,
				Value:    rec.Value,
				Priority: rec.Priority,
			})
		}

		err = d.dnsProv.ConfigureDNS(ctx, domainName, dnsRecords)
		if err != nil {
			log.Errorf("failed to setup DNS automatically: %v", err)
		} else {
			log.Infof("successfully configured DNS records for %s", domainName)

			// 3. Trigger verification check in email provider
			_, verifyErr := d.emailProv.VerifyDomain(ctx, res.DomainID)
			if verifyErr != nil {
				log.Errorf("failed to trigger domain verification: %v", verifyErr)
			}
		}
	}

	// Check current status in email provider
	verified, _ := d.emailProv.VerifyDomain(ctx, res.DomainID)

	existing, err := d.domainRepo.FindByName(domainName)
	if err == nil && existing != nil {
		existing.Verified = verified
		return d.domainRepo.Update(existing)
	}

	domain := &models.Domain{
		Name:     domainName,
		Verified: verified,
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

func (d *domainService) VerifyDomains(ctx context.Context) error {
	domains, err := d.domainRepo.FindAll()
	if err != nil {
		return err
	}

	for _, dom := range domains {
		if !dom.Verified {
			log.Infof("Checking verification status for domain: %s", dom.Name)
			res, err := d.emailProv.RegisterDomain(ctx, dom.Name)
			if err != nil {
				log.Errorf("Worker failed to fetch domain info for %s: %v", dom.Name, err)
				continue
			}

			verified, err := d.emailProv.VerifyDomain(ctx, res.DomainID)
			if err != nil {
				log.Errorf("Worker failed to verify domain status for %s: %v", dom.Name, err)
				continue
			}

			if verified {
				dom.Verified = true
				if updateErr := d.domainRepo.Update(&dom); updateErr != nil {
					log.Errorf("Worker failed to update domain status in DB for %s: %v", dom.Name, updateErr)
				} else {
					log.Infof("Worker successfully verified and activated domain: %s", dom.Name)
				}
			}
		}
	}
	return nil
}


