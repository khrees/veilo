package repositories

import (
	"github.com/khrees/veilo/models"
	"gorm.io/gorm"
)

type DomainRepository interface {
	Create(d *models.Domain) error
	Update(d *models.Domain) error
	Delete(id string) error
	FindAll() ([]models.Domain, error)
	FindByID(id string) (*models.Domain, error)
	FindByName(name string) (*models.Domain, error)
}

type domainRepository struct {
	db *gorm.DB
}

func NewDomainRepository(db *gorm.DB) DomainRepository {
	return &domainRepository{db: db}
}

func (r *domainRepository) Create(d *models.Domain) error {
	return r.db.Create(d).Error
}

func (r *domainRepository) Update(d *models.Domain) error {
	return r.db.Save(d).Error
}

func (r *domainRepository) Delete(id string) error {
	return r.db.Delete(&models.Domain{}, "id = ?", id).Error
}

func (r *domainRepository) FindAll() ([]models.Domain, error) {
	var domains []models.Domain
	return domains, r.db.Find(&domains).Error
}

func (r *domainRepository) FindByID(id string) (*models.Domain, error) {
	var domain models.Domain
	return &domain, r.db.First(&domain, "id = ?", id).Error
}

func (r *domainRepository) FindByName(name string) (*models.Domain, error) {
	var d models.Domain
	return &d, r.db.Where("name = ?", name).First(&d).Error
}
