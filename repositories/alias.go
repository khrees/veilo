package repositories

import (
	"github.com/khrees/cloakee/models"
	"gorm.io/gorm"
)

type AliasRepository interface {
	Create(a *models.Alias) error
	FindAll() ([]models.Alias, error)
	FindByID(id string) (*models.Alias, error)
	Update(id string, updates map[string]any) error
	Delete(id string) error
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

func (r *aliasRepository) FindAll() ([]models.Alias, error) {
	var aliases []models.Alias
	return aliases, r.db.Order("created_at DESC").Find(&aliases).Error
}

func (r *aliasRepository) FindByID(id string) (*models.Alias, error) {
	var alias models.Alias
	return &alias, r.db.First(&alias, "id = ?", id).Error
}

func (r *aliasRepository) Update(id string, updates map[string]any) error {
	return r.db.Model(&models.Alias{}).Where("id = ?", id).Updates(updates).Error
}

func (r *aliasRepository) Delete(id string) error {
	return r.db.Delete(&models.Alias{}, "id = ?", id).Error
}
