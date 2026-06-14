package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/services"
	"github.com/stretchr/testify/assert"
)

func TestAliasService_Create(t *testing.T) {
	t.Parallel()
	mockRepo := &mockAliasRepository{}
	svc := services.NewAliasService(mockRepo)

	input := services.AliasCreateInput{
		Address:   "test@test.com",
		Slug:      "test-slug",
		Domain:    "test.com",
		RealEmail: "real@example.com",
		Enabled:   true,
	}

	alias, err := svc.Create(input)
	assert.NoError(t, err)
	assert.Equal(t, "test@test.com", alias.Address)
	assert.True(t, alias.Enabled)
}

func TestAliasService_GetAll(t *testing.T) {
	t.Parallel()
	mockRepo := &mockAliasRepository{
		aliases: map[string]*models.Alias{
			"1": {ID: uuid.New(), Address: "a1@test.com", Enabled: true},
			"2": {ID: uuid.New(), Address: "a2@test.com", Enabled: false},
		},
	}
	svc := services.NewAliasService(mockRepo)

	aliases, err := svc.GetAll(models.AliasFilter{})
	assert.NoError(t, err)
	assert.Len(t, aliases, 2)
}

func TestAliasService_GetByID(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	mockRepo := &mockAliasRepository{
		aliases: map[string]*models.Alias{
			id.String(): {ID: id, Address: "test@test.com", Enabled: true},
		},
	}
	svc := services.NewAliasService(mockRepo)

	alias, err := svc.GetByID(id.String())
	assert.NoError(t, err)
	assert.Equal(t, "test@test.com", alias.Address)
}

func TestAliasService_FindByAddress(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	mockRepo := &mockAliasRepository{
		aliases: map[string]*models.Alias{
			id.String(): {ID: id, Address: "test@test.com", Enabled: true},
		},
	}
	svc := services.NewAliasService(mockRepo)

	alias, err := svc.FindByAddress("test@test.com")
	assert.NoError(t, err)
	assert.Equal(t, id, alias.ID)
}

func TestAliasService_Update(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	mockRepo := &mockAliasRepository{
		aliases: map[string]*models.Alias{
			id.String(): {ID: id, Address: "test@test.com", RealEmail: "old@example.com", Enabled: true},
		},
	}
	svc := services.NewAliasService(mockRepo)

	err := svc.Update(id.String(), map[string]any{
		"real_email": "new@example.com",
		"enabled":    false,
	})
	assert.NoError(t, err)

	alias, err := mockRepo.FindByID(id.String())
	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", alias.RealEmail)
	assert.False(t, alias.Enabled)
}

func TestAliasService_Delete(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	mockRepo := &mockAliasRepository{
		aliases: map[string]*models.Alias{
			id.String(): {ID: id, Address: "test@test.com", Enabled: true},
		},
	}
	svc := services.NewAliasService(mockRepo)

	err := svc.Delete(id.String())
	assert.NoError(t, err)

	_, err = mockRepo.FindByID(id.String())
	assert.Error(t, err)
}

func TestAliasService_ErrorHandling(t *testing.T) {
	t.Parallel()
	errTest := errors.New("repository error")

	mockRepo := &mockAliasRepository{err: errTest}
	svc := services.NewAliasService(mockRepo)

	input := services.AliasCreateInput{Address: "test@test.com", Slug: "test", Domain: "test.com", RealEmail: "real@example.com"}
	_, err := svc.Create(input)
	assert.ErrorIs(t, err, errTest)

	_, err = svc.GetAll(models.AliasFilter{})
	assert.ErrorIs(t, err, errTest)
}

func TestGenerateSlug(t *testing.T) {
	t.Parallel()
	slug := services.GenerateSlug()
	assert.NotEmpty(t, slug)
	assert.Contains(t, slug, "-")
}

func TestAliasService_DisableExpired(t *testing.T) {
	t.Parallel()
	mockRepo := &mockAliasRepository{}
	svc := services.NewAliasService(mockRepo)

	err := svc.DisableExpired(time.Now())
	assert.NoError(t, err)
}
