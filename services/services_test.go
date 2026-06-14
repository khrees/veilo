package services_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/services"
)

// Mock Domain Repository
type mockDomainRepository struct {
	domains map[string]*models.Domain
	err     error
}

func (m *mockDomainRepository) Create(d *models.Domain) error {
	if m.err != nil {
		return m.err
	}
	d.ID = uuid.New()
	if m.domains == nil {
		m.domains = make(map[string]*models.Domain)
	}
	m.domains[d.ID.String()] = d
	return nil
}

func (m *mockDomainRepository) Delete(id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.domains, id)
	return nil
}

func (m *mockDomainRepository) FindAll() ([]models.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]models.Domain, 0, len(m.domains))
	for _, d := range m.domains {
		result = append(result, *d)
	}
	return result, nil
}

func (m *mockDomainRepository) FindByID(id string) (*models.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}
	d, exists := m.domains[id]
	if !exists {
		return nil, errors.New("domain not found")
	}
	return d, nil
}

func (m *mockDomainRepository) FindByName(name string) (*models.Domain, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, d := range m.domains {
		if d.Name == name {
			return d, nil
		}
	}
	return nil, errors.New("domain not found")
}

// Mock Alias Repository
type mockAliasRepository struct {
	aliases map[string]*models.Alias
	err     error
}

func (m *mockAliasRepository) Create(a *models.Alias) error {
	if m.err != nil {
		return m.err
	}
	a.ID = uuid.New()
	if m.aliases == nil {
		m.aliases = make(map[string]*models.Alias)
	}
	m.aliases[a.ID.String()] = a
	return nil
}

func (m *mockAliasRepository) FindAll(filter models.AliasFilter) ([]models.Alias, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]models.Alias, 0, len(m.aliases))
	for _, a := range m.aliases {
		if filter.Enabled != nil && a.Enabled != *filter.Enabled {
			continue
		}
		if filter.Domain != nil && *filter.Domain != "" && a.Domain != *filter.Domain {
			continue
		}
		result = append(result, *a)
	}
	if filter.Offset != nil && *filter.Offset >= 0 && *filter.Offset < len(result) {
		result = result[*filter.Offset:]
	} else if filter.Offset != nil && *filter.Offset >= len(result) {
		result = []models.Alias{}
	}
	if filter.Limit != nil && *filter.Limit >= 0 && *filter.Limit < len(result) {
		result = result[:*filter.Limit]
	}
	return result, nil
}

func (m *mockAliasRepository) FindByID(id string) (*models.Alias, error) {
	if m.err != nil {
		return nil, m.err
	}
	a, exists := m.aliases[id]
	if !exists {
		return nil, errors.New("alias not found")
	}
	return a, nil
}

func (m *mockAliasRepository) Update(id string, updates map[string]any) error {
	if m.err != nil {
		return m.err
	}
	alias, exists := m.aliases[id]
	if !exists {
		return errors.New("alias not found")
	}
	for key, value := range updates {
		// This is a simplified update for testing
		switch key {
		case "real_email":
			alias.RealEmail = value.(string)
		case "enabled":
			alias.Enabled = value.(bool)
		}
	}
	return nil
}

func (m *mockAliasRepository) Delete(id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.aliases, id)
	return nil
}

func (m *mockAliasRepository) FindByAddress(address string) (*models.Alias, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, a := range m.aliases {
		if a.Address == address {
			return a, nil
		}
	}
	return nil, errors.New("alias not found")
}

// Mock ForwardLog Repository
type mockForwardLogRepository struct {
	logs  map[string]*models.ForwardLog
	stats *models.Stats
	err   error
}

func (m *mockForwardLogRepository) FindByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []models.ForwardLog
	for _, log := range m.logs {
		if log.AliasID.String() == aliasID {
			result = append(result, *log)
		}
	}
	return result, nil
}

func (m *mockForwardLogRepository) GetStats() (*models.Stats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

func (m *mockForwardLogRepository) Create(f *models.ForwardLog) error {
	if m.err != nil {
		return m.err
	}
	f.ID = uuid.New()
	if m.logs == nil {
		m.logs = make(map[string]*models.ForwardLog)
	}
	m.logs[f.ID.String()] = f
	return nil
}

func TestDomainService_Register(t *testing.T) {
	mockRepo := &mockDomainRepository{}
	svc := services.NewDomainService(mockRepo)

	err := svc.Register("test.com")
	if err != nil {
		t.Fatalf("failed to register domain: %v", err)
	}

	domains, err := mockRepo.FindAll()
	if err != nil {
		t.Fatalf("failed to find domains: %v", err)
	}

	if len(domains) != 1 {
		t.Errorf("expected 1 domain, got %d", len(domains))
	}
	if domains[0].Name != "test.com" {
		t.Errorf("expected domain 'test.com', got '%s'", domains[0].Name)
	}
}

func TestDomainService_Remove(t *testing.T) {
	mockRepo := &mockDomainRepository{}
	svc := services.NewDomainService(mockRepo)

	// Register a domain first
	err := svc.Register("test.com")
	if err != nil {
		t.Fatalf("failed to register domain: %v", err)
	}

	// Remove the domain
	err = svc.Remove("test.com")
	if err != nil {
		t.Fatalf("failed to remove domain: %v", err)
	}

	domains, err := mockRepo.FindAll()
	if err != nil {
		t.Fatalf("failed to find domains: %v", err)
	}

	if len(domains) != 0 {
		t.Errorf("expected 0 domains after removal, got %d", len(domains))
	}
}

func TestDomainService_FindAll(t *testing.T) {
	mockRepo := &mockDomainRepository{
		domains: map[string]*models.Domain{
			"1": {ID: uuid.New(), Name: "domain1.com", Verified: true},
			"2": {ID: uuid.New(), Name: "domain2.com", Verified: false},
		},
	}
	svc := services.NewDomainService(mockRepo)

	domains, err := svc.FindAll()
	if err != nil {
		t.Fatalf("failed to find all domains: %v", err)
	}

	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(domains))
	}
}

func TestDomainService_FindByID(t *testing.T) {
	id := uuid.New()
	mockRepo := &mockDomainRepository{
		domains: map[string]*models.Domain{
			id.String(): {ID: id, Name: "test.com", Verified: true},
		},
	}
	svc := services.NewDomainService(mockRepo)

	domain, err := svc.FindByID(id.String())
	if err != nil {
		t.Fatalf("failed to find domain by ID: %v", err)
	}

	if domain.Name != "test.com" {
		t.Errorf("expected domain 'test.com', got '%s'", domain.Name)
	}
}

func TestDomainService_FindByName(t *testing.T) {
	mockRepo := &mockDomainRepository{
		domains: map[string]*models.Domain{
			"1": {Name: "test.com", Verified: true},
		},
	}
	svc := services.NewDomainService(mockRepo)

	domain, err := svc.FindByName("test.com")
	if err != nil {
		t.Fatalf("failed to find domain by name: %v", err)
	}

	if domain.Name != "test.com" {
		t.Errorf("expected domain 'test.com', got '%s'", domain.Name)
	}
}

func TestAliasService_Create(t *testing.T) {
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
	if err != nil {
		t.Fatalf("failed to create alias: %v", err)
	}

	if alias.Address != "test@test.com" {
		t.Errorf("expected address 'test@test.com', got '%s'", alias.Address)
	}
	if alias.Enabled != true {
		t.Errorf("expected enabled true, got %v", alias.Enabled)
	}
}

func TestAliasService_GetAll(t *testing.T) {
	mockRepo := &mockAliasRepository{
		aliases: map[string]*models.Alias{
			"1": {ID: uuid.New(), Address: "a1@test.com", Enabled: true},
			"2": {ID: uuid.New(), Address: "a2@test.com", Enabled: false},
		},
	}
	svc := services.NewAliasService(mockRepo)

	aliases, err := svc.GetAll(models.AliasFilter{})
	if err != nil {
		t.Fatalf("failed to get all aliases: %v", err)
	}

	if len(aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(aliases))
	}
}

func TestAliasService_GetByID(t *testing.T) {
	id := uuid.New()
	mockRepo := &mockAliasRepository{
		aliases: map[string]*models.Alias{
			id.String(): {ID: id, Address: "test@test.com", Enabled: true},
		},
	}
	svc := services.NewAliasService(mockRepo)

	alias, err := svc.GetByID(id.String())
	if err != nil {
		t.Fatalf("failed to get alias by ID: %v", err)
	}

	if alias.Address != "test@test.com" {
		t.Errorf("expected address 'test@test.com', got '%s'", alias.Address)
	}
}

func TestAliasService_Update(t *testing.T) {
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
	if err != nil {
		t.Fatalf("failed to update alias: %v", err)
	}

	alias, err := mockRepo.FindByID(id.String())
	if err != nil {
		t.Fatalf("failed to find alias: %v", err)
	}

	if alias.RealEmail != "new@example.com" {
		t.Errorf("expected real_email 'new@example.com', got '%s'", alias.RealEmail)
	}
	if alias.Enabled != false {
		t.Errorf("expected enabled false, got %v", alias.Enabled)
	}
}

func TestAliasService_Delete(t *testing.T) {
	id := uuid.New()
	mockRepo := &mockAliasRepository{
		aliases: map[string]*models.Alias{
			id.String(): {ID: id, Address: "test@test.com", Enabled: true},
		},
	}
	svc := services.NewAliasService(mockRepo)

	err := svc.Delete(id.String())
	if err != nil {
		t.Fatalf("failed to delete alias: %v", err)
	}

	_, err = mockRepo.FindByID(id.String())
	if err == nil {
		t.Error("expected error when finding deleted alias")
	}
}

func TestForwardLogService_GetByAliasID(t *testing.T) {
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
	if err != nil {
		t.Fatalf("failed to get forward logs: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("expected 3 forward logs, got %d", len(logs))
	}
}

func TestForwardLogService_GetStats(t *testing.T) {
	mockRepo := &mockForwardLogRepository{
		stats: &models.Stats{
			TotalAliases:   10,
			TotalForwarded: 100,
			TotalBlocked:   5,
		},
	}
	svc := services.NewForwardLogService(mockRepo)

	stats, err := svc.GetStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalAliases != 10 {
		t.Errorf("expected 10 total aliases, got %d", stats.TotalAliases)
	}
	if stats.TotalForwarded != 100 {
		t.Errorf("expected 100 total forwarded, got %d", stats.TotalForwarded)
	}
	if stats.TotalBlocked != 5 {
		t.Errorf("expected 5 total blocked, got %d", stats.TotalBlocked)
	}
}

// Test error handling scenarios
func TestDomainService_ErrorHandling(t *testing.T) {
	errTest := errors.New("repository error")

	mockRepo := &mockDomainRepository{err: errTest}
	svc := services.NewDomainService(mockRepo)

	// Test Register with error
	err := svc.Register("test.com")
	if !errors.Is(err, errTest) {
		t.Errorf("expected error '%v', got '%v'", errTest, err)
	}

	// Test FindAll with error
	_, err = svc.FindAll()
	if !errors.Is(err, errTest) {
		t.Errorf("expected error '%v', got '%v'", errTest, err)
	}
}

func TestAliasService_ErrorHandling(t *testing.T) {
	errTest := errors.New("repository error")

	mockRepo := &mockAliasRepository{err: errTest}
	svc := services.NewAliasService(mockRepo)

	// Test Create with error
	input := services.AliasCreateInput{Address: "test@test.com", Slug: "test", Domain: "test.com", RealEmail: "real@example.com"}
	_, err := svc.Create(input)
	if !errors.Is(err, errTest) {
		t.Errorf("expected error '%v', got '%v'", errTest, err)
	}

	// Test GetAll with error
	_, err = svc.GetAll(models.AliasFilter{})
	if !errors.Is(err, errTest) {
		t.Errorf("expected error '%v', got '%v'", errTest, err)
	}
}

func TestForwardLogService_ErrorHandling(t *testing.T) {
	errTest := errors.New("repository error")

	mockRepo := &mockForwardLogRepository{err: errTest}
	svc := services.NewForwardLogService(mockRepo)

	// Test GetByAliasID with error
	_, err := svc.GetByAliasID(uuid.New().String(), 10, 0)
	if !errors.Is(err, errTest) {
		t.Errorf("expected error '%v', got '%v'", errTest, err)
	}

	// Test GetStats with error
	_, err = svc.GetStats()
	if !errors.Is(err, errTest) {
		t.Errorf("expected error '%v', got '%v'", errTest, err)
	}
}
