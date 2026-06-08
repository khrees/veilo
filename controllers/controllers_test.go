package controllers_test

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
	"github.com/khrees/veilo/controllers"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
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

func (m *mockDomainService) FindByName(name string) (*models.Domain, error) {
	args := m.Called(name)
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

func (m *mockAliasService) GetAll(filter models.AliasFilter) ([]models.Alias, error) {
	args := m.Called(filter)
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
func createTestApp(deps controllers.RouteDeps) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	controllers.SetupRoutes(app, deps)
	return app
}

func TestDomainController_RegisterDomain(t *testing.T) {
	mockSvc := new(mockDomainService)
	mockSvc.On("Register", "test.com").Return(nil)

	app := createTestApp(controllers.RouteDeps{DomainSvc: mockSvc})

	body := map[string]string{"domain": "test.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/v1/domains", bytes.NewBuffer(jsonBody))
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

	app := createTestApp(controllers.RouteDeps{DomainSvc: mockSvc})

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/v1/domains", bytes.NewBufferString("{invalid json"))
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
		{ID: uuid.New(), Name: "domain1.com", Verified: true},
		{ID: uuid.New(), Name: "domain2.com", Verified: false},
	}
	mockSvc.On("FindAll").Return(domains, nil)

	app := createTestApp(controllers.RouteDeps{DomainSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/domains", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var apiResp struct {
		Success bool            `json:"success"`
		Data    []models.Domain `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Errorf("expected success true, got false")
	}

	if len(apiResp.Data) != 2 {
		t.Errorf("expected 2 domains, got %d", len(apiResp.Data))
	}

	mockSvc.AssertExpectations(t)
}

func TestDomainController_ListDomains_Error(t *testing.T) {
	mockSvc := new(mockDomainService)
	mockSvc.On("FindAll").Return(nil, errors.New("boom"))

	app := createTestApp(controllers.RouteDeps{DomainSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/domains", nil)

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
	domain := &models.Domain{ID: uuid.New(), Name: "test.com", Verified: true}
	mockSvc.On("FindByName", "test.com").Return(domain, nil)

	app := createTestApp(controllers.RouteDeps{DomainSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/domains/test.com", nil)

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

	app := createTestApp(controllers.RouteDeps{DomainSvc: mockSvc})

	req := httptest.NewRequest(http.MethodDelete, "/v1/domains/test.com", nil)

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

	app := createTestApp(controllers.RouteDeps{AliasSvc: mockSvc})

	body := map[string]any{
		"address":    "test@test.com",
		"slug":       "test-slug",
		"domain":     "test.com",
		"real_email": "real@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/v1/aliases", bytes.NewBuffer(jsonBody))
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

func TestAliasController_CreateAlias_Error(t *testing.T) {
	mockSvc := new(mockAliasService)
	mockSvc.On("Create", mock.Anything).Return((*models.Alias)(nil), errors.New("boom"))

	app := createTestApp(controllers.RouteDeps{AliasSvc: mockSvc})

	body := map[string]any{
		"address":    "test@test.com",
		"slug":       "test-slug",
		"domain":     "test.com",
		"real_email": "real@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/v1/aliases", bytes.NewBuffer(jsonBody))
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

	app := createTestApp(controllers.RouteDeps{AliasSvc: mockSvc})

	body := map[string]any{
		"address":    "test@test.com",
		"slug":       "test-slug",
		"domain":     "test.com",
		"real_email": "real@example.com",
		"label":      "test-label",
		"enabled":    false,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/v1/aliases", bytes.NewBuffer(jsonBody))
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

func TestAliasController_ListAliases(t *testing.T) {
	mockSvc := new(mockAliasService)
	aliases := []models.Alias{
		{ID: uuid.New(), Address: "a1@test.com", Enabled: true},
		{ID: uuid.New(), Address: "a2@test.com", Enabled: false},
	}
	mockSvc.On("GetAll", mock.Anything).Return(aliases, nil)

	app := createTestApp(controllers.RouteDeps{AliasSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/aliases", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var apiResp struct {
		Success bool           `json:"success"`
		Data    []models.Alias `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Errorf("expected success true, got false")
	}

	if len(apiResp.Data) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(apiResp.Data))
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_ListAliases_Filters(t *testing.T) {
	mockSvc := new(mockAliasService)

	// We expect multiple calls to GetAll with different filters
	// 1. enabled=true
	enabledTrue := true
	mockSvc.On("GetAll", models.AliasFilter{Enabled: &enabledTrue}).Return([]models.Alias{}, nil).Once()

	// 2. domain=cooldomain.xyz
	domainVal := "cooldomain.xyz"
	mockSvc.On("GetAll", models.AliasFilter{Domain: &domainVal}).Return([]models.Alias{}, nil).Once()

	// 3. limit=50&offset=10
	limitVal := 50
	offsetVal := 10
	mockSvc.On("GetAll", models.AliasFilter{Limit: &limitVal, Offset: &offsetVal}).Return([]models.Alias{}, nil).Once()

	app := createTestApp(controllers.RouteDeps{AliasSvc: mockSvc})

	// Test 1: enabled=true
	req1 := httptest.NewRequest(http.MethodGet, "/v1/aliases?enabled=true", nil)
	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	// Test 2: domain=cooldomain.xyz
	req2 := httptest.NewRequest(http.MethodGet, "/v1/aliases?domain=cooldomain.xyz", nil)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp2.StatusCode)
	}

	// Test 3: limit=50&offset=10
	req3 := httptest.NewRequest(http.MethodGet, "/v1/aliases?limit=50&offset=10", nil)
	resp3, err := app.Test(req3)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp3.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestAliasController_GetAlias(t *testing.T) {
	mockSvc := new(mockAliasService)
	alias := &models.Alias{ID: uuid.New(), Address: "test@test.com", Enabled: true}
	mockSvc.On("GetByID", mock.Anything).Return(alias, nil)

	app := createTestApp(controllers.RouteDeps{AliasSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/aliases/some-id", nil)

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

	app := createTestApp(controllers.RouteDeps{AliasSvc: mockSvc})

	body := map[string]any{
		"real_email": "new@example.com",
		"enabled":    false,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/v1/aliases/some-id", bytes.NewBuffer(jsonBody))
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

func TestAliasController_DeleteAlias(t *testing.T) {
	mockSvc := new(mockAliasService)
	mockSvc.On("Delete", mock.Anything).Return(nil)

	app := createTestApp(controllers.RouteDeps{AliasSvc: mockSvc})

	req := httptest.NewRequest(http.MethodDelete, "/v1/aliases/some-id", nil)

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

func TestForwardLogController_GetForwardLogs(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	aliasID := uuid.New()
	logs := []models.ForwardLog{
		{AliasID: aliasID, Direction: "inbound", Status: "delivered"},
		{AliasID: aliasID, Direction: "inbound", Status: "blocked"},
	}
	mockSvc.On("GetByAliasID", aliasID.String(), 50, 0).Return(logs, nil)

	app := createTestApp(controllers.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/forward-logs/"+aliasID.String(), nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var apiResp struct {
		Success bool                `json:"success"`
		Data    []models.ForwardLog `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Errorf("expected success true, got false")
	}

	if len(apiResp.Data) != 2 {
		t.Errorf("expected 2 forward logs, got %d", len(apiResp.Data))
	}

	mockSvc.AssertExpectations(t)
}

func TestForwardLogController_GetForwardLogs_WithPagination(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	mockSvc.On("GetByAliasID", mock.Anything, mock.Anything, mock.Anything).Return([]models.ForwardLog{}, nil)

	app := createTestApp(controllers.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/forward-logs/"+uuid.New().String()+"?limit=10&offset=20", nil)

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

	app := createTestApp(controllers.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/forward-logs/"+uuid.New().String(), nil)

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

func TestForwardLogController_GetForwardLogs_AliasEndpoint(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	aliasID := uuid.New()
	mockSvc.On("GetByAliasID", aliasID.String(), 10, 5).Return([]models.ForwardLog{}, nil)

	app := createTestApp(controllers.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/aliases/"+aliasID.String()+"/logs?limit=10&offset=5", nil)

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

	app := createTestApp(controllers.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/stats", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var apiResp struct {
		Success bool               `json:"success"`
		Data    repositories.Stats `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Errorf("expected success true, got false")
	}

	if apiResp.Data.TotalAliases != 10 {
		t.Errorf("expected 10 total aliases, got %d", apiResp.Data.TotalAliases)
	}
	if apiResp.Data.TotalForwarded != 100 {
		t.Errorf("expected 100 total forwarded, got %d", apiResp.Data.TotalForwarded)
	}
	if apiResp.Data.TotalBlocked != 5 {
		t.Errorf("expected 5 total blocked, got %d", apiResp.Data.TotalBlocked)
	}

	mockSvc.AssertExpectations(t)
}

func TestStatsController_GetStats_Error(t *testing.T) {
	mockSvc := new(mockForwardLogService)
	mockSvc.On("GetStats").Return((*repositories.Stats)(nil), errors.New("boom"))

	app := createTestApp(controllers.RouteDeps{ForwardLogSvc: mockSvc})

	req := httptest.NewRequest(http.MethodGet, "/v1/stats", nil)

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
