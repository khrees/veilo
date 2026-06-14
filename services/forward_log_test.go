package services_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/services"
	"github.com/stretchr/testify/assert"
)

func TestForwardLogService_GetByAliasID(t *testing.T) {
	t.Parallel()
	aliasID := uuid.New()
	mockRepo := &mockForwardLogRepository{
		logs: map[string]*models.ForwardLog{
			"1": {AliasID: aliasID, Direction: "inbound", Status: "delivered"},
			"2": {AliasID: aliasID, Direction: "inbound", Status: "blocked"},
			"3": {AliasID: aliasID, Direction: "reply", Status: "delivered"},
		},
	}
	svc := services.NewForwardLogService(mockRepo)

	logs, err := svc.GetByAliasID(aliasID.String(), 10, 0)
	assert.NoError(t, err)
	assert.Len(t, logs, 3)
}

func TestForwardLogService_Create(t *testing.T) {
	t.Parallel()
	aliasID := uuid.New()
	mockRepo := &mockForwardLogRepository{}
	svc := services.NewForwardLogService(mockRepo)

	logEntry := &models.ForwardLog{
		AliasID:   aliasID,
		Direction: "inbound",
		Status:    "delivered",
	}

	err := svc.Create(logEntry)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, logEntry.ID)
}

func TestForwardLogService_GetStats(t *testing.T) {
	t.Parallel()
	mockRepo := &mockForwardLogRepository{
		stats: &models.Stats{
			TotalAliases:   10,
			TotalForwarded: 100,
			TotalBlocked:   5,
		},
	}
	svc := services.NewForwardLogService(mockRepo)

	stats, err := svc.GetStats()
	assert.NoError(t, err)
	assert.Equal(t, int64(10), stats.TotalAliases)
	assert.Equal(t, int64(100), stats.TotalForwarded)
	assert.Equal(t, int64(5), stats.TotalBlocked)
}

func TestForwardLogService_ErrorHandling(t *testing.T) {
	t.Parallel()
	errTest := errors.New("repository error")

	mockRepo := &mockForwardLogRepository{err: errTest}
	svc := services.NewForwardLogService(mockRepo)

	_, err := svc.GetByAliasID(uuid.New().String(), 10, 0)
	assert.ErrorIs(t, err, errTest)

	_, err = svc.GetStats()
	assert.ErrorIs(t, err, errTest)
}
