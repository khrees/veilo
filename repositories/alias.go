// Package repositories contains the data access layer for the application.
package repositories

import (
	"time"

	"github.com/khrees/veilo/models"
	"gorm.io/gorm"
)

type AliasRepository interface {
	Create(a *models.Alias) error
	FindAll(filter models.AliasFilter) ([]models.Alias, error)
	FindByID(id string) (*models.Alias, error)
	FindByAddress(address string) (*models.Alias, error)
	Update(id string, updates map[string]any) error
	Delete(id string) error
	DisableExpired(now time.Time) error
}

type aliasRepository struct {
	db *gorm.DB
}

func NewAliasRepository(db *gorm.DB) AliasRepository {
	return &aliasRepository{db: db}
}

func (r *aliasRepository) Create(a *models.Alias) error {
	return r.db.Create(a).Error
}

func (r *aliasRepository) FindAll(filter models.AliasFilter) ([]models.Alias, error) {
	var aliases []models.Alias
	query := r.db.Model(&models.Alias{}).Order("created_at DESC")
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if filter.Domain != nil && *filter.Domain != "" {
		query = query.Where("domain = ?", *filter.Domain)
	}
	if filter.Limit != nil {
		query = query.Limit(*filter.Limit)
	}
	if filter.Offset != nil {
		query = query.Offset(*filter.Offset)
	}
	return aliases, query.Find(&aliases).Error
}

func (r *aliasRepository) FindByID(id string) (*models.Alias, error) {
	var alias models.Alias
	return &alias, r.db.First(&alias, "id = ?", id).Error
}

func (r *aliasRepository) FindByAddress(address string) (*models.Alias, error) {
	var alias models.Alias
	return &alias, r.db.First(&alias, "address = ?", address).Error
}

func (r *aliasRepository) Update(id string, updates map[string]any) error {
	return r.db.Model(&models.Alias{}).Where("id = ?", id).Updates(updates).Error
}

func (r *aliasRepository) Delete(id string) error {
	return r.db.Delete(&models.Alias{}, "id = ?", id).Error
}

func (r *aliasRepository) DisableExpired(now time.Time) error {
	return r.db.Model(&models.Alias{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND enabled = ?", now, true).
		Update("enabled", false).Error
}
