package repositories

import (
	"github.com/khrees/cloakee/models"
	"gorm.io/gorm"
)

func CreateDomain(db *gorm.DB, d *models.Domain) error {
	return db.Create(d).Error
}

func DeleteDomain(db *gorm.DB, id string) error {
	return db.Delete(&models.Domain{}, "id = ?", id).Error
}

func FindAllDomains(db *gorm.DB) ([]models.Domain, error) {
	var domains []models.Domain
	return domains, db.Find(&domains).Error
}

func FindDomainByID(db *gorm.DB, id string) (*models.Domain, error) {
	var domain models.Domain
	return &domain, db.First(&domain, "id = ?", id).Error
}
