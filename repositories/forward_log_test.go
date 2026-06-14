package repositories_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
	"github.com/stretchr/testify/assert"
)

func TestForwardLogRepository_FindByAliasID(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.ForwardLog{})

	repo := repositories.NewForwardLogRepository(db)

	aliasID := uuid.New()

	// Insert forward logs using repo
	forwardLogs := []*models.ForwardLog{
		{AliasID: aliasID, Direction: "inbound", Status: "delivered", CreatedAt: time.Now().Add(-2 * time.Hour)},
		{AliasID: aliasID, Direction: "inbound", Status: "blocked", CreatedAt: time.Now().Add(-1 * time.Hour)},
		{AliasID: aliasID, Direction: "reply", Status: "delivered", CreatedAt: time.Now()},
	}

	for _, log := range forwardLogs {
		err := repo.Create(log)
		assert.NoError(t, err)
	}

	// Find by alias ID with pagination
	result, err := repo.FindByAliasID(aliasID.String(), 2, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestForwardLogRepository_GetStats(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{}, &models.ForwardLog{})

	repo := repositories.NewForwardLogRepository(db)

	// Create some aliases
	aliasRepo := repositories.NewAliasRepository(db)
	aliases := []*models.Alias{
		{Address: "a1@test.com", Slug: "a1", Domain: "test.com", RealEmail: "r1@example.com", Enabled: true},
		{Address: "a2@test.com", Slug: "a2", Domain: "test.com", RealEmail: "r2@example.com", Enabled: true},
		{Address: "a3@test.com", Slug: "a3", Domain: "test.com", RealEmail: "r3@example.com", Enabled: false},
	}
	for _, a := range aliases {
		err := aliasRepo.Create(a)
		assert.NoError(t, err)
	}

	// Create forward logs with different statuses
	aliasID := aliases[0].ID
	forwardLogs := []*models.ForwardLog{
		{AliasID: aliasID, Direction: "inbound", Status: "delivered"},
		{AliasID: aliasID, Direction: "inbound", Status: "delivered"},
		{AliasID: aliasID, Direction: "inbound", Status: "blocked"},
		{AliasID: aliasID, Direction: "reply", Status: "bounced"},
	}
	for _, log := range forwardLogs {
		err := repo.Create(log)
		assert.NoError(t, err)
	}

	// Get stats
	stats, err := repo.GetStats()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), stats.TotalAliases)
	assert.Equal(t, int64(2), stats.TotalForwarded)
	assert.Equal(t, int64(1), stats.TotalBlocked)
}
