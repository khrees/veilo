package repositories_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
	"github.com/stretchr/testify/assert"
)

func TestDomainRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	domain := &models.Domain{
		Name:     "test.com",
		Verified: false,
	}

	err := repo.Create(domain)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, domain.ID)

	// Test Update
	domain.Verified = true
	err = repo.Update(domain)
	assert.NoError(t, err)

	// Test FindByName
	found, err := repo.FindByName("test.com")
	assert.NoError(t, err)
	assert.True(t, found.Verified)
}

func TestDomainRepository_Delete(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	domain := &models.Domain{
		Name:     "test.com",
		Verified: false,
	}
	err := repo.Create(domain)
	assert.NoError(t, err)

	// Delete the domain
	err = repo.Delete(domain.ID.String())
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(domain.ID.String())
	assert.Error(t, err)
}

func TestDomainRepository_FindAll(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	domains := []*models.Domain{
		{Name: "domain1.com", Verified: false},
		{Name: "domain2.com", Verified: true},
		{Name: "domain3.com", Verified: false},
	}

	for _, d := range domains {
		err := repo.Create(d)
		assert.NoError(t, err)
	}

	// Find all
	result, err := repo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestDomainRepository_FindByID(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	domain := &models.Domain{
		Name:     "test.com",
		Verified: false,
	}
	err := repo.Create(domain)
	assert.NoError(t, err)

	// Find by ID
	result, err := repo.FindByID(domain.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, domain.ID, result.ID)
	assert.Equal(t, "test.com", result.Name)
}

func TestDomainRepository_FindByName(t *testing.T) {
	t.Parallel()
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	domain := &models.Domain{
		Name:     "test.com",
		Verified: false,
	}
	err := repo.Create(domain)
	assert.NoError(t, err)

	// Find by domain name
	result, err := repo.FindByName("test.com")
	assert.NoError(t, err)
	assert.Equal(t, "test.com", result.Name)
}
