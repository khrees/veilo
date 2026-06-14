package repositories_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
	"github.com/stretchr/testify/assert"
)

func TestReplyTokenRepository_All(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.ReplyToken{})

	repo := repositories.NewReplyTokenRepository(db)

	aliasID := uuid.New()
	subject := "Hello test"
	thread := "msg_123"

	// 1. Create
	token := &models.ReplyToken{
		Token:           "test-token-1",
		AliasID:         aliasID,
		OriginalSender:  "sender@example.com",
		OriginalSubject: &subject,
		ThreadID:        &thread,
		ExpiresAt:       time.Now().Add(-1 * time.Hour), // Already expired
	}

	token2 := &models.ReplyToken{
		Token:           "test-token-2",
		AliasID:         aliasID,
		OriginalSender:  "sender2@example.com",
		OriginalSubject: &subject,
		ThreadID:        &thread,
		ExpiresAt:       time.Now().Add(1 * time.Hour), // Active
	}

	err := repo.Create(token)
	assert.NoError(t, err)
	err = repo.Create(token2)
	assert.NoError(t, err)

	// 2. FindByToken
	found, err := repo.FindByToken("test-token-2")
	assert.NoError(t, err)
	assert.Equal(t, "sender2@example.com", found.OriginalSender)

	// 3. DeleteExpired
	err = repo.DeleteExpired(time.Now())
	assert.NoError(t, err)

	// Verify test-token-1 (expired) is gone
	_, err = repo.FindByToken("test-token-1")
	assert.Error(t, err)

	// Verify test-token-2 (active) is still there
	_, err = repo.FindByToken("test-token-2")
	assert.NoError(t, err)

	// 4. Delete
	err = repo.Delete("test-token-2")
	assert.NoError(t, err)

	_, err = repo.FindByToken("test-token-2")
	assert.Error(t, err)
}
