package repositories_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	// Use SQLite for testing - in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Create tables manually without PostgreSQL-specific defaults
	err = db.Exec(`CREATE TABLE "domains" (
		"id" TEXT PRIMARY KEY,
		"name" TEXT NOT NULL UNIQUE,
		"verified" INTEGER NOT NULL DEFAULT 0,
		"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"deleted_at" DATETIME
	)`).Error
	if err != nil {
		panic(err)
	}

	err = db.Exec(`CREATE TABLE "aliases" (
		"id" TEXT PRIMARY KEY,
		"address" TEXT NOT NULL UNIQUE,
		"slug" TEXT NOT NULL,
		"domain" TEXT NOT NULL,
		"real_email" TEXT NOT NULL,
		"display_name" TEXT,
		"label" TEXT,
		"enabled" INTEGER NOT NULL DEFAULT 1,
		"forward_count" INTEGER NOT NULL DEFAULT 0,
		"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"last_used_at" DATETIME
	)`).Error
	if err != nil {
		panic(err)
	}

	err = db.Exec(`CREATE TABLE "reply_tokens" (
		"token" TEXT PRIMARY KEY,
		"alias_id" TEXT NOT NULL,
		"original_sender" TEXT NOT NULL,
		"original_subject" TEXT,
		"thread_id" TEXT,
		"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"expires_at" DATETIME NOT NULL
	)`).Error
	if err != nil {
		panic(err)
	}

	err = db.Exec(`CREATE TABLE "forward_logs" (
		"id" TEXT PRIMARY KEY,
		"alias_id" TEXT NOT NULL,
		"direction" TEXT NOT NULL,
		"sender" TEXT,
		"subject" TEXT,
		"status" TEXT NOT NULL,
		"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error
	if err != nil {
		panic(err)
	}

	return db
}

func TestDomainRepository_Create(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	domain := &models.Domain{
		Name:     "test.com",
		Verified: false,
	}

	err := repo.Create(domain)
	if err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}

	if domain.ID == uuid.Nil {
		t.Error("expected domain to have a non-nil UUID after creation")
	}

	// Test Update
	domain.Verified = true
	if err := repo.Update(domain); err != nil {
		t.Fatalf("failed to update domain: %v", err)
	}

	// Test FindByName
	found, err := repo.FindByName("test.com")
	if err != nil {
		t.Fatalf("failed to find domain by name: %v", err)
	}
	if !found.Verified {
		t.Error("expected domain to be verified after update")
	}
}

func TestDomainRepository_Delete(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	// Create a domain first
	domain := &models.Domain{
		Name:     "test.com",
		Verified: false,
	}
	err := repo.Create(domain)
	if err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}

	// Delete the domain
	err = repo.Delete(domain.ID.String())
	if err != nil {
		t.Fatalf("failed to delete domain: %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID(domain.ID.String())
	if err == nil {
		t.Error("expected error when finding deleted domain")
	}
}

func TestDomainRepository_FindAll(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	// Create multiple domains
	domains := []*models.Domain{
		{Name: "domain1.com", Verified: false},
		{Name: "domain2.com", Verified: true},
		{Name: "domain3.com", Verified: false},
	}

	for _, d := range domains {
		err := repo.Create(d)
		if err != nil {
			t.Fatalf("failed to create domain: %v", err)
		}
	}

	// Find all
	result, err := repo.FindAll()
	if err != nil {
		t.Fatalf("failed to find all domains: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 domains, got %d", len(result))
	}
}

func TestDomainRepository_FindByID(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	// Create a domain
	domain := &models.Domain{
		Name:     "test.com",
		Verified: false,
	}
	err := repo.Create(domain)
	if err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}

	// Find by ID
	result, err := repo.FindByID(domain.ID.String())
	if err != nil {
		t.Fatalf("failed to find domain by ID: %v", err)
	}

	if result.ID != domain.ID {
		t.Errorf("expected ID %v, got %v", domain.ID, result.ID)
	}
	if result.Name != "test.com" {
		t.Errorf("expected domain 'test.com', got '%s'", result.Name)
	}
}

func TestDomainRepository_FindByName(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Domain{})

	repo := repositories.NewDomainRepository(db)

	// Create a domain
	domain := &models.Domain{
		Name:     "test.com",
		Verified: false,
	}
	err := repo.Create(domain)
	if err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}

	// Find by domain name
	result, err := repo.FindByName("test.com")
	if err != nil {
		t.Fatalf("failed to find domain by name: %v", err)
	}

	if result.Name != "test.com" {
		t.Errorf("expected domain 'test.com', got '%s'", result.Name)
	}
}

func TestAliasRepository_Create(t *testing.T) {
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
	if err != nil {
		t.Fatalf("failed to create alias: %v", err)
	}

	if alias.ID == uuid.Nil {
		t.Error("expected alias to have a non-nil UUID after creation")
	}
}

func TestAliasRepository_FindAll(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	// Create aliases
	aliases := []*models.Alias{
		{Address: "a1@test.com", Slug: "a1", Domain: "test.com", RealEmail: "r1@example.com", Enabled: true},
		{Address: "a2@test.com", Slug: "a2", Domain: "test.com", RealEmail: "r2@example.com", Enabled: false},
		{Address: "a3@other.com", Slug: "a3", Domain: "other.com", RealEmail: "r3@example.com", Enabled: true},
	}

	for _, a := range aliases {
		err := repo.Create(a)
		if err != nil {
			t.Fatalf("failed to create alias: %v", err)
		}
	}

	// Find all
	result, err := repo.FindAll(models.AliasFilter{})
	if err != nil {
		t.Fatalf("failed to find all aliases: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 aliases, got %d", len(result))
	}

	// Filter by Enabled = true
	enabledTrue := true
	result, err = repo.FindAll(models.AliasFilter{Enabled: &enabledTrue})
	if err != nil {
		t.Fatalf("failed to filter enabled: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 enabled aliases, got %d", len(result))
	}

	// Filter by Enabled = false
	enabledFalse := false
	result, err = repo.FindAll(models.AliasFilter{Enabled: &enabledFalse})
	if err != nil {
		t.Fatalf("failed to filter disabled: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 disabled alias, got %d", len(result))
	}

	// Filter by Domain = "test.com"
	domainFilter := "test.com"
	result, err = repo.FindAll(models.AliasFilter{Domain: &domainFilter})
	if err != nil {
		t.Fatalf("failed to filter domain: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 aliases for test.com, got %d", len(result))
	}

	// Test pagination: limit = 2, offset = 1
	limitVal := 2
	offsetVal := 1
	result, err = repo.FindAll(models.AliasFilter{Limit: &limitVal, Offset: &offsetVal})
	if err != nil {
		t.Fatalf("failed to paginate: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 paged aliases, got %d", len(result))
	}
}

func TestAliasRepository_FindByID(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	// Create an alias
	alias := &models.Alias{
		Address:   "test@test.com",
		Slug:      "test-slug",
		Domain:    "test.com",
		RealEmail: "real@example.com",
		Enabled:   true,
	}
	err := repo.Create(alias)
	if err != nil {
		t.Fatalf("failed to create alias: %v", err)
	}

	// Find by ID
	result, err := repo.FindByID(alias.ID.String())
	if err != nil {
		t.Fatalf("failed to find alias by ID: %v", err)
	}

	if result.Address != "test@test.com" {
		t.Errorf("expected address 'test@test.com', got '%s'", result.Address)
	}

	// Find by Address
	resultByAddr, err := repo.FindByAddress("test@test.com")
	if err != nil {
		t.Fatalf("failed to find alias by address: %v", err)
	}

	if resultByAddr.ID != alias.ID {
		t.Errorf("expected ID %s, got %s", alias.ID, resultByAddr.ID)
	}
}

func TestAliasRepository_Update(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	// Create an alias
	alias := &models.Alias{
		Address:   "test@test.com",
		Slug:      "test-slug",
		Domain:    "test.com",
		RealEmail: "real@example.com",
		Enabled:   true,
	}
	err := repo.Create(alias)
	if err != nil {
		t.Fatalf("failed to create alias: %v", err)
	}

	// Update
	err = repo.Update(alias.ID.String(), map[string]any{
		"real_email": "new@example.com",
		"enabled":    false,
	})
	if err != nil {
		t.Fatalf("failed to update alias: %v", err)
	}

	// Verify
	result, err := repo.FindByID(alias.ID.String())
	if err != nil {
		t.Fatalf("failed to find alias: %v", err)
	}

	if result.RealEmail != "new@example.com" {
		t.Errorf("expected real_email 'new@example.com', got '%s'", result.RealEmail)
	}
	if result.Enabled != false {
		t.Errorf("expected enabled false, got %v", result.Enabled)
	}
}

func TestAliasRepository_Delete(t *testing.T) {
	db := setupTestDB()
	defer db.Migrator().DropTable(&models.Alias{})

	repo := repositories.NewAliasRepository(db)

	// Create an alias
	alias := &models.Alias{
		Address:   "test@test.com",
		Slug:      "test-slug",
		Domain:    "test.com",
		RealEmail: "real@example.com",
		Enabled:   true,
	}
	err := repo.Create(alias)
	if err != nil {
		t.Fatalf("failed to create alias: %v", err)
	}

	// Delete
	err = repo.Delete(alias.ID.String())
	if err != nil {
		t.Fatalf("failed to delete alias: %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID(alias.ID.String())
	if err == nil {
		t.Error("expected error when finding deleted alias")
	}
}

func TestForwardLogRepository_FindByAliasID(t *testing.T) {
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
		if err != nil {
			t.Fatalf("failed to create forward log: %v", err)
		}
	}

	// Find by alias ID with pagination
	result, err := repo.FindByAliasID(aliasID.String(), 2, 0)
	if err != nil {
		t.Fatalf("failed to find forward logs: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 forward logs, got %d", len(result))
	}
}

func TestForwardLogRepository_GetStats(t *testing.T) {
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
		aliasRepo.Create(a)
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
		if err != nil {
			t.Fatalf("failed to create forward log: %v", err)
		}
	}

	// Get stats
	stats, err := repo.GetStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalAliases != 3 {
		t.Errorf("expected 3 total aliases, got %d", stats.TotalAliases)
	}
	if stats.TotalForwarded != 2 {
		t.Errorf("expected 2 total forwarded, got %d", stats.TotalForwarded)
	}
	if stats.TotalBlocked != 1 {
		t.Errorf("expected 1 total blocked, got %d", stats.TotalBlocked)
	}
}

func TestReplyTokenRepository_All(t *testing.T) {
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

	if err := repo.Create(token); err != nil {
		t.Fatalf("failed to create token: %v", err)
	}
	if err := repo.Create(token2); err != nil {
		t.Fatalf("failed to create token2: %v", err)
	}

	// 2. FindByToken
	found, err := repo.FindByToken("test-token-2")
	if err != nil {
		t.Fatalf("failed to find token: %v", err)
	}
	if found.OriginalSender != "sender2@example.com" {
		t.Errorf("expected sender2@example.com, got %s", found.OriginalSender)
	}

	// 3. DeleteExpired
	if err := repo.DeleteExpired(time.Now()); err != nil {
		t.Fatalf("failed to delete expired tokens: %v", err)
	}

	// Verify test-token-1 (expired) is gone
	_, err = repo.FindByToken("test-token-1")
	if err == nil {
		t.Error("expected expired token to be deleted, but it was found")
	}

	// Verify test-token-2 (active) is still there
	_, err = repo.FindByToken("test-token-2")
	if err != nil {
		t.Errorf("expected active token to remain, got err: %v", err)
	}

	// 4. Delete
	if err := repo.Delete("test-token-2"); err != nil {
		t.Fatalf("failed to delete token: %v", err)
	}

	_, err = repo.FindByToken("test-token-2")
	if err == nil {
		t.Error("expected token to be deleted")
	}
}
