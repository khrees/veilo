package controllers

import (
	"testing"
)

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

	if bounced.Bounce.SubType != "Suppressed" {
		t.Errorf("expected bounce subType 'Suppressed', got '%s'", bounced.Bounce.SubType)
	}

	if bounced.Tags.Category != "confirm_email" {
		t.Errorf("expected tags category 'confirm_email', got '%s'", bounced.Tags.Category)
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

	if email.Subject != "Hello test" {
		t.Errorf("expected subject 'Hello test', got '%s'", email.Subject)
	}
}
