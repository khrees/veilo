package repositories

import (
	"github.com/khrees/cloakee/models"
	"gorm.io/gorm"
)

type ForwardLogRepository interface {
	FindByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error)
	GetStats() (*Stats, error)
}

type forwardLogRepository struct {
	db *gorm.DB
}

func NewForwardLogRepository(db *gorm.DB) ForwardLogRepository {
	return &forwardLogRepository{db: db}
}

func (r *forwardLogRepository) FindByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error) {
	var logs []models.ForwardLog
	return logs, r.db.Where("alias_id = ?", aliasID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
}

func (r *forwardLogRepository) GetStats() (*Stats, error) {
	var s Stats

	if err := r.db.Model(&models.Alias{}).Count(&s.TotalAliases).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&models.ForwardLog{}).
		Where("status = ?", "delivered").
		Count(&s.TotalForwarded).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&models.ForwardLog{}).
		Where("status = ?", "blocked").
		Count(&s.TotalBlocked).Error; err != nil {
		return nil, err
	}

	return &s, nil
}
