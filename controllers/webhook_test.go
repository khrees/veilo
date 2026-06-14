package controllers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
	"github.com/stretchr/testify/mock"
)

// ---------- Mock Webhook Service ----------

type mockWebhookSvc struct {
	mock.Mock
}

func (m *mockWebhookSvc) ProcessEmailReceived(ctx context.Context, input services.EmailReceivedInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// ---------- Helper functions ----------

func generateTestSvixHeaders(secret string, payload []byte) map[string]string {
	msgID := "msg_test123"
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	secretPart := strings.TrimPrefix(secret, "whsec_")
	key, err := base64.StdEncoding.DecodeString(secretPart)
	if err != nil {
		panic(err)
	}
	toSign := fmt.Sprintf("%s.%s.%s", msgID, timestamp, string(payload))
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(toSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return map[string]string{
		"svix-id":        msgID,
		"svix-timestamp": timestamp,
		"svix-signature": fmt.Sprintf("v1,%s", signature),
	}
}

// ---------- Tests ----------

func TestParseWebhookEvent_EmailBounced(t *testing.T) {
	payload := `{
  "type": "email.bounced",
  "created_at": "2026-11-22T23:41:12.126Z",
  "data": {
    "broadcast_id": "8b146471-e88e-4322-86af-016cd36fd216",
    "created_at": "2026-11-22T23:41:11.894719+00:00",
    "email_id": "56761188-7520-42d8-8898-ff6fc54ce618",
    "from": "Acme <onboarding@resend.dev>",
    "to": ["delivered@resend.dev"],
    "subject": "Sending this example",
    "template_id": "43f68331-0622-4e15-8202-246a0388854b",
    "bounce": {
      "message": "The recipient's email address is on the suppression list because it has a recent history of producing hard bounces.",
      "subType": "Suppressed",
      "type": "Permanent"
    },
    "tags": {
      "category": "confirm_email"
    }
  }
}`

	webhook, err := ParseWebhookEvent([]byte(payload))
	if err != nil {
		t.Fatalf("failed to parse webhook: %v", err)
	}

	if webhook.Type != "email.bounced" {
		t.Errorf("expected type 'email.bounced', got '%s'", webhook.Type)
	}

	expectedTime := "2026-11-22T23:41:12.126Z"
	if webhook.CreatedAt.Format("2006-01-02T15:04:05.000Z") != expectedTime {
		t.Errorf("expected created_at '%s', got '%s'", expectedTime, webhook.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	}

	bounced, ok := webhook.Data.(EmailBounced)
	if !ok {
		t.Fatalf("expected data type EmailBounced, got %T", webhook.Data)
	}

	if bounced.BroadcastID != "8b146471-e88e-4322-86af-016cd36fd216" {
		t.Errorf("expected broadcast_id '8b146471-e88e-4322-86af-016cd36fd216', got '%s'", bounced.BroadcastID)
	}
}

func TestParseWebhookEvent_EmailReceived(t *testing.T) {
	payload := `{
  "type": "email.received",
  "created_at": "2026-11-22T23:41:12.126Z",
  "data": {
    "email_id": "56761188-7520-42d8-8898-ff6fc54ce618",
    "created_at": "2026-11-22T23:41:11.894719+00:00",
    "from": "Acme <onboarding@resend.dev>",
    "to": ["delivered@resend.dev"],
    "message_id": "123",
    "subject": "Hello test"
  }
}`

	webhook, err := ParseWebhookEvent([]byte(payload))
	if err != nil {
		t.Fatalf("failed to parse webhook: %v", err)
	}

	if webhook.Type != "email.received" {
		t.Errorf("expected type 'email.received', got '%s'", webhook.Type)
	}

	email, ok := webhook.Data.(EmailReceived)
	if !ok {
		t.Fatalf("expected data type EmailReceived, got %T", webhook.Data)
	}

	if email.EmailID != "56761188-7520-42d8-8898-ff6fc54ce618" {
		t.Errorf("expected email_id '56761188-7520-42d8-8898-ff6fc54ce618', got '%s'", email.EmailID)
	}
}

func TestWebhookController_HandleInboundWebhook_InvalidSignature(t *testing.T) {
	app := fiber.New()
	secret := "whsec_dGVzdF9zZWNyZXRfd2hzZWNsZW5ndGhfaXNfZW5vdWdo"
	controller := NewWebhookController(RouteDeps{
		WebhookSecret: secret,
	})
	controller.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodPost, "/webhook/inbound", strings.NewReader(`{}`))
	req.Header.Set("svix-signature", "v1,invalidsig")
	req.Header.Set("svix-id", "msg_123")
	req.Header.Set("svix-timestamp", "123456")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to execute test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestWebhookController_HandleInboundWebhook_Success(t *testing.T) {
	app := fiber.New()
	secret := "whsec_dGVzdF9zZWNyZXRfd2hzZWNsZW5ndGhfaXNfZW5vdWdo"

	mockSvc := new(mockWebhookSvc)

	deps := RouteDeps{
		WebhookSvc:    mockSvc,
		WebhookSecret: secret,
	}

	controller := NewWebhookController(deps)
	controller.RegisterRoutes(app)

	// Inbound Webhook Payload
	payload := `{
		"type": "email.received",
		"created_at": "2026-06-12T15:00:00Z",
		"data": {
			"email_id": "email_123",
			"from": "John Doe <john@example.com>",
			"to": ["alias@cooldomain.xyz"],
			"subject": "Hello test",
			"message_id": "msg_id_original"
		}
	}`

	expectedInput := services.EmailReceivedInput{
		EmailID:   "email_123",
		From:      "John Doe <john@example.com>",
		To:        []string{"alias@cooldomain.xyz"},
		Subject:   "Hello test",
		MessageID: "msg_id_original",
	}

	mockSvc.On("ProcessEmailReceived", mock.Anything, expectedInput).Return(nil)

	// Prepare Request
	req := httptest.NewRequest(http.MethodPost, "/webhook/inbound", strings.NewReader(payload))
	headers := generateTestSvixHeaders(secret, []byte(payload))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}

func TestWebhookController_HandleInboundWebhook_ServiceError(t *testing.T) {
	app := fiber.New()
	secret := "whsec_dGVzdF9zZWNyZXRfd2hzZWNsZW5ndGhfaXNfZW5vdWdo"

	mockSvc := new(mockWebhookSvc)

	deps := RouteDeps{
		WebhookSvc:    mockSvc,
		WebhookSecret: secret,
	}

	controller := NewWebhookController(deps)
	controller.RegisterRoutes(app)

	payload := `{
		"type": "email.received",
		"created_at": "2026-06-12T15:00:00Z",
		"data": {
			"email_id": "email_123",
			"from": "John Doe <john@example.com>",
			"to": ["alias@cooldomain.xyz"],
			"subject": "Hello test"
		}
	}`

	mockSvc.On("ProcessEmailReceived", mock.Anything, mock.Anything).Return(errors.New("service failure"))

	req := httptest.NewRequest(http.MethodPost, "/webhook/inbound", strings.NewReader(payload))
	headers := generateTestSvixHeaders(secret, []byte(payload))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}

	mockSvc.AssertExpectations(t)
}
