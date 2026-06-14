package repositories_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
	"github.com/stretchr/testify/assert"
)

func TestAliasRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	alias := &models.Alias{
		Address:      "test@test.com",
		Slug:         "test-slug",
		Domain:       "test.com",
		RealEmail:    "real@example.com",
		Enabled:      true,
		ForwardCount: 0,
	}

	err := repo.Create(alias)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, alias.ID)
}

func TestAliasRepository_FindAll(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	aliases := []*models.Alias{
		{Address: "a1@test.com", Slug: "a1", Domain: "test.com", RealEmail: "r1@example.com", Enabled: true},
		{Address: "a2@test.com", Slug: "a2", Domain: "test.com", RealEmail: "r2@example.com", Enabled: false},
		{Address: "a3@other.com", Slug: "a3", Domain: "other.com", RealEmail: "r3@example.com", Enabled: true},
	}

	for _, a := range aliases {
		err := repo.Create(a)
		assert.NoError(t, err)
	}

	// Find all
	result, err := repo.FindAll(models.AliasFilter{})
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Filter by Enabled = true
	enabledTrue := true
	result, err = repo.FindAll(models.AliasFilter{Enabled: &enabledTrue})
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Filter by Enabled = false
	enabledFalse := false
	result, err = repo.FindAll(models.AliasFilter{Enabled: &enabledFalse})
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Filter by Domain = "test.com"
	domainFilter := "test.com"
	result, err = repo.FindAll(models.AliasFilter{Domain: &domainFilter})
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test pagination: limit = 2, offset = 1
	limitVal := 2
	offsetVal := 1
	result, err = repo.FindAll(models.AliasFilter{Limit: &limitVal, Offset: &offsetVal})
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestAliasRepository_FindByID(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	alias := &models.Alias{
		Address:   "test@test.com",
		Slug:      "test-slug",
		Domain:    "test.com",
		RealEmail: "real@example.com",
		Enabled:   true,
	}
	err := repo.Create(alias)
	assert.NoError(t, err)

	// Find by ID
	result, err := repo.FindByID(alias.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "test@test.com", result.Address)

	// Find by Address
	resultByAddr, err := repo.FindByAddress("test@test.com")
	assert.NoError(t, err)
	assert.Equal(t, alias.ID, resultByAddr.ID)
}

func TestAliasRepository_Update(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	alias := &models.Alias{
		Address:   "test@test.com",
		Slug:      "test-slug",
		Domain:    "test.com",
		RealEmail: "real@example.com",
		Enabled:   true,
	}
	err := repo.Create(alias)
	assert.NoError(t, err)

	// Update
	err = repo.Update(alias.ID.String(), map[string]any{
		"real_email": "new@example.com",
		"enabled":    false,
	})
	assert.NoError(t, err)

	// Verify
	result, err := repo.FindByID(alias.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", result.RealEmail)
	assert.False(t, result.Enabled)
}

func TestAliasRepository_Delete(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	alias := &models.Alias{
		Address:   "test@test.com",
		Slug:      "test-slug",
		Domain:    "test.com",
		RealEmail: "real@example.com",
		Enabled:   true,
	}
	err := repo.Create(alias)
	assert.NoError(t, err)

	// Delete
	err = repo.Delete(alias.ID.String())
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(alias.ID.String())
	assert.Error(t, err)
}
