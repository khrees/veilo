package services

import (
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
)

// ForwardLogService interface for forward log operations
type ForwardLogService interface {
	Create(f *models.ForwardLog) error
	GetByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error)
	GetStats() (*models.Stats, error)
}

// forwardLogService implements ForwardLogService
type forwardLogService struct {
	forwardLogRepo repositories.ForwardLogRepository
}

// NewForwardLogService will instantiate ForwardLogService
func NewForwardLogService(forwardLogRepo repositories.ForwardLogRepository) ForwardLogService {
	return &forwardLogService{
		forwardLogRepo: forwardLogRepo,
	}
}

func (f *forwardLogService) Create(log *models.ForwardLog) error {
	return f.forwardLogRepo.Create(log)
}

// GetByAliasID returns forward logs for a specific alias
func (f *forwardLogService) GetByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error) {
	return f.forwardLogRepo.FindByAliasID(aliasID, limit, offset)
}

// GetStats returns statistics
func (f *forwardLogService) GetStats() (*models.Stats, error) {
	return f.forwardLogRepo.GetStats()
}


