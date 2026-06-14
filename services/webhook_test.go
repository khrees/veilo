package services_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/services"
	"github.com/khrees/veilo/providers"
	"github.com/resend/resend-go/v3"
	"github.com/stretchr/testify/mock"
)

// ---------- Mock Repositories ----------

type mockAliasRepo struct {
	mock.Mock
}

func (m *mockAliasRepo) Create(a *models.Alias) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *mockAliasRepo) FindAll(filter models.AliasFilter) ([]models.Alias, error) {
	args := m.Called(filter)
	return args.Get(0).([]models.Alias), args.Error(1)
}

func (m *mockAliasRepo) FindByID(id string) (*models.Alias, error) {
	args := m.Called(id)
	if a := args.Get(0); a != nil {
		return a.(*models.Alias), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAliasRepo) FindByAddress(address string) (*models.Alias, error) {
	args := m.Called(address)
	if a := args.Get(0); a != nil {
		return a.(*models.Alias), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAliasRepo) Update(id string, updates map[string]any) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *mockAliasRepo) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockForwardLogRepo struct {
	mock.Mock
}

func (m *mockForwardLogRepo) Create(f *models.ForwardLog) error {
	args := m.Called(f)
	return args.Error(0)
}

func (m *mockForwardLogRepo) FindByAliasID(aliasID string, limit, offset int) ([]models.ForwardLog, error) {
	args := m.Called(aliasID, limit, offset)
	return args.Get(0).([]models.ForwardLog), args.Error(1)
}

func (m *mockForwardLogRepo) GetStats() (*models.Stats, error) {
	args := m.Called()
	if a := args.Get(0); a != nil {
		return a.(*models.Stats), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockReplyTokenRepo struct {
	mock.Mock
}

func (m *mockReplyTokenRepo) Create(t *models.ReplyToken) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *mockReplyTokenRepo) FindByToken(token string) (*models.ReplyToken, error) {
	args := m.Called(token)
	if r := args.Get(0); r != nil {
		return r.(*models.ReplyToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockReplyTokenRepo) Delete(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

// ---------- Tests ----------

func TestWebhookService_ProcessEmailReceived_ForwardFlow_Success(t *testing.T) {
	// Setup mock Resend server
	resendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/emails/receiving/") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "email_123",
				"subject": "Hello test",
				"html": "<p>Hello</p>",
				"text": "Hello"
			}`))
		} else if strings.Contains(r.URL.Path, "/emails") && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "sent_123"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer resendServer.Close()

	// Mock Repos
	mockAlias := new(mockAliasRepo)
	mockForwardLog := new(mockForwardLogRepo)
	mockReplyToken := new(mockReplyTokenRepo)

	resendClient := resend.NewClient("re_test123")
	parsedURL, _ := url.Parse(resendServer.URL)
	resendClient.BaseURL = parsedURL

	svc := services.NewWebhookService(mockAlias, mockForwardLog, mockReplyToken, providers.NewResendEmailProvider(resendClient), 90)

	// Inputs
	input := services.EmailReceivedInput{
		EmailID:   "email_123",
		From:      "John Doe <john@example.com>",
		To:        []string{"alias@cooldomain.xyz"},
		Subject:   "Hello test",
		MessageID: "msg_id_original",
	}

	aliasUUID := uuid.New()
	alias := &models.Alias{
		ID:           aliasUUID,
		Address:      "alias@cooldomain.xyz",
		Domain:       "cooldomain.xyz",
		RealEmail:    "real@gmail.com",
		Enabled:      true,
		ForwardCount: 5,
	}

	// Mock Expectations
	mockAlias.On("FindByAddress", "alias@cooldomain.xyz").Return(alias, nil)
	mockReplyToken.On("Create", mock.MatchedBy(func(t *models.ReplyToken) bool {
		return t.AliasID == aliasUUID && t.OriginalSender == "john@example.com"
	})).Return(nil)
	mockAlias.On("Update", aliasUUID.String(), mock.MatchedBy(func(updates map[string]any) bool {
		return updates["forward_count"] == 6
	})).Return(nil)
	mockForwardLog.On("Create", mock.MatchedBy(func(log *models.ForwardLog) bool {
		return log.AliasID == aliasUUID && log.Direction == "inbound" && log.Status == "delivered"
	})).Return(nil)

	err := svc.ProcessEmailReceived(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	mockAlias.AssertExpectations(t)
	mockForwardLog.AssertExpectations(t)
	mockReplyToken.AssertExpectations(t)
}

func TestWebhookService_ProcessEmailReceived_ForwardFlow_AliasNotFound(t *testing.T) {
	mockAlias := new(mockAliasRepo)
	mockForwardLog := new(mockForwardLogRepo)
	mockReplyToken := new(mockReplyTokenRepo)

	svc := services.NewWebhookService(mockAlias, mockForwardLog, mockReplyToken, nil, 90)

	input := services.EmailReceivedInput{
		EmailID: "email_123",
		From:    "john@example.com",
		To:      []string{"notfound@cooldomain.xyz"},
		Subject: "Hello test",
	}

	mockAlias.On("FindByAddress", "notfound@cooldomain.xyz").Return((*models.Alias)(nil), errors.New("not found"))

	err := svc.ProcessEmailReceived(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error (silent drop), got %v", err)
	}

	mockAlias.AssertExpectations(t)
}

func TestWebhookService_ProcessEmailReceived_ForwardFlow_AliasDisabled(t *testing.T) {
	mockAlias := new(mockAliasRepo)
	mockForwardLog := new(mockForwardLogRepo)
	mockReplyToken := new(mockReplyTokenRepo)

	svc := services.NewWebhookService(mockAlias, mockForwardLog, mockReplyToken, nil, 90)

	input := services.EmailReceivedInput{
		EmailID: "email_123",
		From:    "john@example.com",
		To:      []string{"disabled@cooldomain.xyz"},
		Subject: "Hello test",
	}

	aliasUUID := uuid.New()
	alias := &models.Alias{
		ID:      aliasUUID,
		Address: "disabled@cooldomain.xyz",
		Enabled: false,
	}

	mockAlias.On("FindByAddress", "disabled@cooldomain.xyz").Return(alias, nil)
	mockForwardLog.On("Create", mock.MatchedBy(func(log *models.ForwardLog) bool {
		return log.AliasID == aliasUUID && log.Direction == "inbound" && log.Status == "blocked"
	})).Return(nil)

	err := svc.ProcessEmailReceived(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error (silent drop), got %v", err)
	}

	mockAlias.AssertExpectations(t)
	mockForwardLog.AssertExpectations(t)
}

func TestWebhookService_ProcessEmailReceived_ReplyFlow_Success(t *testing.T) {
	resendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/emails/receiving/") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "email_reply",
				"subject": "Re: Hello test",
				"html": "<p>Reply content</p>",
				"text": "Reply content"
			}`))
		} else if strings.Contains(r.URL.Path, "/emails") && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "sent_reply_123"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer resendServer.Close()

	mockAlias := new(mockAliasRepo)
	mockForwardLog := new(mockForwardLogRepo)
	mockReplyToken := new(mockReplyTokenRepo)

	resendClient := resend.NewClient("re_test123")
	parsedURL, _ := url.Parse(resendServer.URL)
	resendClient.BaseURL = parsedURL

	svc := services.NewWebhookService(mockAlias, mockForwardLog, mockReplyToken, providers.NewResendEmailProvider(resendClient), 90)

	input := services.EmailReceivedInput{
		EmailID: "email_reply",
		From:    "real@gmail.com",
		To:      []string{"reply+token_abc_123@cooldomain.xyz"},
		Subject: "Re: Hello test",
	}

	aliasUUID := uuid.New()
	replyToken := &models.ReplyToken{
		Token:          "token_abc_123",
		AliasID:        aliasUUID,
		OriginalSender: "john@example.com",
		ExpiresAt:      time.Now().Add(1 * time.Hour),
	}

	alias := &models.Alias{
		ID:      aliasUUID,
		Address: "alias@cooldomain.xyz",
		Enabled: true,
	}

	mockReplyToken.On("FindByToken", "token_abc_123").Return(replyToken, nil)
	mockAlias.On("FindByID", aliasUUID.String()).Return(alias, nil)
	mockForwardLog.On("Create", mock.MatchedBy(func(log *models.ForwardLog) bool {
		return log.AliasID == aliasUUID && log.Direction == "reply" && log.Status == "delivered"
	})).Return(nil)

	err := svc.ProcessEmailReceived(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	mockReplyToken.AssertExpectations(t)
	mockAlias.AssertExpectations(t)
	mockForwardLog.AssertExpectations(t)
}

func TestWebhookService_ProcessEmailReceived_ReplyFlow_ExpiredToken(t *testing.T) {
	mockAlias := new(mockAliasRepo)
	mockForwardLog := new(mockForwardLogRepo)
	mockReplyToken := new(mockReplyTokenRepo)

	svc := services.NewWebhookService(mockAlias, mockForwardLog, mockReplyToken, nil, 90)

	input := services.EmailReceivedInput{
		EmailID: "email_reply",
		From:    "real@gmail.com",
		To:      []string{"reply+expired_token@cooldomain.xyz"},
		Subject: "Re: Hello test",
	}

	aliasUUID := uuid.New()
	replyToken := &models.ReplyToken{
		Token:          "expired_token",
		AliasID:        aliasUUID,
		OriginalSender: "john@example.com",
		ExpiresAt:      time.Now().Add(-1 * time.Hour), // Expired!
	}

	mockReplyToken.On("FindByToken", "expired_token").Return(replyToken, nil)

	err := svc.ProcessEmailReceived(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	mockReplyToken.AssertExpectations(t)
}
