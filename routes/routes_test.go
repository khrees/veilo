package routes_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
	"github.com/khrees/veilo/routes"
	"github.com/khrees/veilo/services"
	"github.com/stretchr/testify/mock"
)

// Mock service interfaces
type mockDomainService struct {
	mock.Mock
}

func (m *mockDomainService) Register(domainName string) error {
	args := m.Called(domainName)
	return args.Error(0)
}

func (m *mockDomainService) Remove(domainName string) error {
	args := m.Called(domainName)
	return args.Error(0)
}

func (m *mockDomainService) FindAll() ([]models.Domain, error) {
	args := m.Called()
	if arg0 := args.Get(0); arg0 != nil {
		return arg0.([]models.Domain), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockDomainService) FindByID(id string) (*models.Domain, error) {
	args := m.Called(id)
	if arg0 := args.Get(0); arg0 != nil {
		return arg0.(*models.Domain), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockDomainService) FindByDomain(domain string) (*models.Domain, error) {
	args := m.Called(domain)
	if arg0 := args.Get(0); arg0 != nil {
		return arg0.(*models.Domain), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockAliasService struct {
	mock.Mock
}

func (m *mockAliasService) Create(input services.AliasCreateInput) (*models.Alias, error) {
	args := m.Called(input)
	if arg0 := args.Get(0); arg0 != nil {
		return arg0.(*models.Alias), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAliasService) GetAll() ([]models.Alias, error) {
	args := m.Called()
	if arg0 := args.Get(0); arg0 != nil {
		return arg0.([]models.Alias), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAliasService) GetByID(id string) (*models.Alias, error) {
	args := m.Called(id)
	if arg0 := args.Get(0); arg0 != nil {
		return arg0.(*models.Alias), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAliasService) Update(id string, updates map[string]any) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *mockAliasService) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockForwardLogService struct {
	mock.Mock
}

func (m *mockForwardLogService) GetByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error) {
	args := m.Called(aliasID, limit, offset)
	if arg0 := args.Get(0); arg0 != nil {
		return arg0.([]models.ForwardLog), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockForwardLogService) GetStats() (*repositories.Stats, error) {
	args := m.Called()
	if arg0 := args.Get(0); arg0 != nil {
		return arg0.(*repositories.Stats), args.Error(1)
	}
	return nil, args.Error(1)
}

// Helper to create test app
func createTestApp(deps routes.RouteDeps) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	routes.SetupRoutes(app, deps)
	return app
}

func TestDomainController_RegisterDomain(t *testing.T) {
	mockSvc := new(mockDomainService)
	mockSvc.On("Register", "test.com").Return(nil)

	app := createTestApp(routes.RouteDeps{DomainSvc: mockSvc})

	body := map[string]string{"domain": "test.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestDomainController_RegisterDomain_ValidationError(t *testing.T) {
	mockSvc := new(mockDomainService)

	app := createTestApp(routes.RouteDeps{DomainSvc: mockSvc})

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/domains", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d for validation error, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestDomainController_ListDomains(t *testing.T) {
	mockSvc := new(mockDomainService)
	domains := []models.Domain{
		{ID: uuid.New(), Domain: "domain1.com", Verified: true},
		{ID: uuid.New(), Domain: "domain2.com", Verified: false},
	}
	mockSvc.On("FindAll").Return(domains, nil)

	app := createTestApp(routes.RouteDeps{DomainSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var result []models.Domain
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 domains, got %d", len(result))
	}

	mockSvc.AssertExpectations(t)
}

func TestDomainController_ListDomains_Error(t *testing.T) {
	mockSvc := new(mockDomainService)
	mockSvc.On("FindAll").Return(nil, errors.New("boom"))

	app := createTestApp(routes.RouteDeps{DomainSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/domains", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestDomainController_GetDomain(t *testing.T) {
	mockSvc := new(mockDomainService)
	domain := &models.Domain{ID: uuid.New(), Domain: "test.com", Verified: true}
	mockSvc.On("FindByDomain", "test.com").Return(domain, nil)

	app := createTestApp(routes.RouteDeps{DomainSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/domains/test.com", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestDomainController_RemoveDomain(t *testing.T) {
	mockSvc := new(mockDomainService)
	mockSvc.On("Remove", "test.com").Return(nil)

	app := createTestApp(routes.RouteDeps{DomainSvc: mockSvc})

	req := httptest.NewRequest(http.MethodDelete, "/api/domains/test.com", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_CreateAlias(t *testing.T) {
	mockSvc := new(mockAliasService)
	alias := &models.Alias{
		ID:           uuid.New(),
		Address:      "test@test.com",
		Slug:         "test-slug",
		Domain:       "test.com",
		RealEmail:    "real@example.com",
		Enabled:      true,
		ForwardCount: 0,
	}
	mockSvc.On("Create", mock.Anything).Return(alias, nil)

	app := createTestApp(routes.RouteDeps{AliasSvc: mockSvc})

	body := map[string]any{
		"address":    "test@test.com",
		"slug":       "test-slug",
		"domain":     "test.com",
		"real_email": "real@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_CreateAlias_Error(t *testing.T) {
	mockSvc := new(mockAliasService)
	mockSvc.On("Create", mock.Anything).Return((*models.Alias)(nil), errors.New("boom"))

	app := createTestApp(routes.RouteDeps{AliasSvc: mockSvc})

	body := map[string]any{
		"address":    "test@test.com",
		"slug":       "test-slug",
		"domain":     "test.com",
		"real_email": "real@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_CreateAlias_WithOptionalFields(t *testing.T) {
	mockSvc := new(mockAliasService)
	alias := &models.Alias{
		ID:           uuid.New(),
		Address:      "test@test.com",
		Slug:         "test-slug",
		Domain:       "test.com",
		RealEmail:    "real@example.com",
		Label:        stringPtr("test-label"),
		Enabled:      false,
		ForwardCount: 0,
	}
	mockSvc.On("Create", mock.Anything).Return(alias, nil)

	app := createTestApp(routes.RouteDeps{AliasSvc: mockSvc})

	body := map[string]any{
		"address":    "test@test.com",
		"slug":       "test-slug",
		"domain":     "test.com",
		"real_email": "real@example.com",
		"label":      "test-label",
		"enabled":    false,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_ListAliases(t *testing.T) {
	mockSvc := new(mockAliasService)
	aliases := []models.Alias{
		{ID: uuid.New(), Address: "a1@test.com", Enabled: true},
		{ID: uuid.New(), Address: "a2@test.com", Enabled: false},
	}
	mockSvc.On("GetAll").Return(aliases, nil)

	app := createTestApp(routes.RouteDeps{AliasSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/aliases", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var result []models.Alias
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(result))
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_GetAlias(t *testing.T) {
	mockSvc := new(mockAliasService)
	alias := &models.Alias{ID: uuid.New(), Address: "test@test.com", Enabled: true}
	mockSvc.On("GetByID", mock.Anything).Return(alias, nil)

	app := createTestApp(routes.RouteDeps{AliasSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/aliases/some-id", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_UpdateAlias(t *testing.T) {
	mockSvc := new(mockAliasService)
	mockSvc.On("Update", mock.Anything, mock.Anything).Return(nil)

	app := createTestApp(routes.RouteDeps{AliasSvc: mockSvc})

	body := map[string]any{
		"real_email": "new@example.com",
		"enabled":    false,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/aliases/some-id", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_DeleteAlias(t *testing.T) {
	mockSvc := new(mockAliasService)
	mockSvc.On("Delete", mock.Anything).Return(nil)

	app := createTestApp(routes.RouteDeps{AliasSvc: mockSvc})

	req := httptest.NewRequest(http.MethodDelete, "/api/aliases/some-id", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestForwardLogController_GetForwardLogs(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	aliasID := uuid.New()
	logs := []models.ForwardLog{
		{AliasID: aliasID, Direction: "inbound", Status: "delivered"},
		{AliasID: aliasID, Direction: "inbound", Status: "blocked"},
	}
	mockSvc.On("GetByAliasID", aliasID.String(), 50, 0).Return(logs, nil)

	app := createTestApp(routes.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/forward-logs/"+aliasID.String(), nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var result []models.ForwardLog
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 forward logs, got %d", len(result))
	}

	mockSvc.AssertExpectations(t)
}

func TestForwardLogController_GetForwardLogs_WithPagination(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	mockSvc.On("GetByAliasID", mock.Anything, mock.Anything, mock.Anything).Return([]models.ForwardLog{}, nil)

	app := createTestApp(routes.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/forward-logs/"+uuid.New().String()+"?limit=10&offset=20", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestForwardLogController_GetForwardLogs_DefaultPagination(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	mockSvc.On("GetByAliasID", mock.Anything, 50, 0).Return([]models.ForwardLog{}, nil)

	app := createTestApp(routes.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/forward-logs/"+uuid.New().String(), nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestStatsController_GetStats(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	stats := &repositories.Stats{
		TotalAliases:   10,
		TotalForwarded: 100,
		TotalBlocked:   5,
	}
	mockSvc.On("GetStats").Return(stats, nil)

	app := createTestApp(routes.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var result repositories.Stats
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.TotalAliases != 10 {
		t.Errorf("expected 10 total aliases, got %d", result.TotalAliases)
	}
	if result.TotalForwarded != 100 {
		t.Errorf("expected 100 total forwarded, got %d", result.TotalForwarded)
	}
	if result.TotalBlocked != 5 {
		t.Errorf("expected 5 total blocked, got %d", result.TotalBlocked)
	}

	mockSvc.AssertExpectations(t)
}

func TestStatsController_GetStats_Error(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	mockSvc.On("GetStats").Return((*repositories.Stats)(nil), errors.New("boom"))

	app := createTestApp(routes.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
