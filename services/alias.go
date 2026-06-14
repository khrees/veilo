// Package services implements the business logic and core services for the Veilo application.
package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
)

// AliasService interface for alias operations
type AliasService interface {
	Create(input AliasCreateInput) (*models.Alias, error)
	GetAll(filter models.AliasFilter) ([]models.Alias, error)
	GetByID(id string) (*models.Alias, error)
	FindByAddress(address string) (*models.Alias, error)
	Update(id string, updates map[string]any) error
	Delete(id string) error
	DisableExpired(now time.Time) error
}

// AliasCreateInput groups the values needed to create an alias.
type AliasCreateInput struct {
	Address     string
	Slug        string
	Domain      string
	RealEmail   string
	DisplayName *string
	Label       *string
	Enabled     bool
	ExpiresAt   *time.Time
	MaxForwards *int
}

var adjectives = []string{
	"glowing", "radiant", "whispering", "silent", "frosty", "golden",
	"silver", "crimson", "azure", "mystic", "shadowy", "stellar",
	"cosmic", "wild", "gentle", "bouncy", "jolly", "merry", "speedy",
	"vibrant", "serene", "dusk", "dawn", "misty", "stormy", "cloudy",
}

var nouns = []string{
	"umbrella", "forest", "sunset", "river", "mountain", "ocean",
	"breeze", "galaxy", "comet", "nebula", "meadow", "canyon",
	"beacon", "harbor", "castle", "fortress", "glade", "oasis",
	"pioneer", "valley", "summit", "island", "desert", "tundra",
}

// GenerateSlug creates a random human-readable slug like "cosmic-nebula-042".
func GenerateSlug() string {
	adjIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	if err != nil {
		panic(fmt.Sprintf("crypto/rand failure: %v", err))
	}
	nounIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	if err != nil {
		panic(fmt.Sprintf("crypto/rand failure: %v", err))
	}
	num, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		panic(fmt.Sprintf("crypto/rand failure: %v", err))
	}

	slug := fmt.Sprintf("%s-%s-%03d", adjectives[adjIdx.Int64()], nouns[nounIdx.Int64()], num.Int64())
	if len(slug) > 25 {
		return slug[:25]
	}
	return slug
}

// allowedUpdateFields is the whitelist of fields that can be updated via the API.
var allowedUpdateFields = map[string]bool{
	"address": true, "real_email": true, "display_name": true, "label": true, "enabled": true,
	"expires_at": true, "max_forwards": true,
}

// aliasService implements AliasService
type aliasService struct {
	aliasRepo repositories.AliasRepository
}

// NewAliasService will instantiate AliasService
func NewAliasService(aliasRepo repositories.AliasRepository) AliasService {
	return &aliasService{
		aliasRepo: aliasRepo,
	}
}

// Create creates a new alias
func (a *aliasService) Create(input AliasCreateInput) (*models.Alias, error) {
	// Extract slug from address if slug is empty
	if input.Slug == "" && input.Address != "" {
		parts := strings.Split(input.Address, "@")
		if len(parts) > 0 {
			input.Slug = parts[0]
		}
	}
	if input.Slug == "" {
		input.Slug = GenerateSlug()
	}
	if input.Address == "" {
		input.Address = fmt.Sprintf("%s@%s", input.Slug, input.Domain)
	}

	alias := &models.Alias{
		Address:      input.Address,
		Slug:         input.Slug,
		Domain:       input.Domain,
		RealEmail:    input.RealEmail,
		DisplayName:  input.DisplayName,
		Label:        input.Label,
		Enabled:      input.Enabled,
		ForwardCount: 0,
		ExpiresAt:    input.ExpiresAt,
		MaxForwards:  input.MaxForwards,
	}

	err := a.aliasRepo.Create(alias)
	if err != nil {
		return nil, err
	}

	return alias, nil
}

// GetAll returns all aliases
func (a *aliasService) GetAll(filter models.AliasFilter) ([]models.Alias, error) {
	return a.aliasRepo.FindAll(filter)
}

// GetByID returns an alias by ID
func (a *aliasService) GetByID(id string) (*models.Alias, error) {
	return a.aliasRepo.FindByID(id)
}

// FindByAddress returns an alias by address
func (a *aliasService) FindByAddress(address string) (*models.Alias, error) {
	return a.aliasRepo.FindByAddress(address)
}

// Update modifies an existing alias
func (a *aliasService) Update(id string, updates map[string]any) error {
	filtered := make(map[string]any, len(updates))
	for k, v := range updates {
		if allowedUpdateFields[k] {
			filtered[k] = v
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return a.aliasRepo.Update(id, filtered)
}

// Delete removes an alias
func (a *aliasService) Delete(id string) error {
	return a.aliasRepo.Delete(id)
}

func (a *aliasService) DisableExpired(now time.Time) error {
	return a.aliasRepo.DisableExpired(now)
}
