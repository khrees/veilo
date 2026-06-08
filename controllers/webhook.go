package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	svix "github.com/svix/svix-webhooks/go"
)

// WebhookEvent represents a Resend webhook event with a type-discriminated Data field.
type WebhookEvent struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Data      any       `json:"data"`
}

// ParseWebhookEvent unmarshals a raw JSON webhook payload into a WebhookEvent,
// discriminating on "type" to deserialise the "data" field into the correct Go type.
func ParseWebhookEvent(raw []byte) (WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return event, err
	}

	switch event.Type {
	case "email.received":
		var payload struct {
			Data EmailReceived `json:"data"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return event, err
		}
		event.Data = payload.Data

	case "email.bounced":
		var payload struct {
			Data EmailBounced `json:"data"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return event, err
		}
		event.Data = payload.Data
	}

	return event, nil
}

// ---------- Typed data structs ----------

type EmailReceived struct {
	EmailID     string        `json:"email_id"`
	CreatedAt   time.Time     `json:"created_at"`
	From        string        `json:"from"`
	To          []string      `json:"to"`
	Bcc         []interface{} `json:"bcc"`
	Cc          []interface{} `json:"cc"`
	MessageID   string        `json:"message_id"`
	Subject     string        `json:"subject"`
	Attachments []Attachment  `json:"attachments"`
}

type Attachment struct {
	ID                 string `json:"id"`
	Filename           string `json:"filename"`
	ContentType        string `json:"content_type"`
	ContentDisposition string `json:"content_disposition"`
	ContentID          string `json:"content_id"`
}

type EmailBounced struct {
	BroadcastID string    `json:"broadcast_id"`
	CreatedAt   time.Time `json:"created_at"`
	EmailID     string    `json:"email_id"`
	From        string    `json:"from"`
	To          []string  `json:"to"`
	Subject     string    `json:"subject"`
	TemplateID  string    `json:"template_id"`
	Bounce      struct {
		Message string `json:"message"`
		SubType string `json:"subType"`
		Type    string `json:"type"`
	} `json:"bounce"`
	Tags struct {
		Category string `json:"category"`
	} `json:"tags"`
}

// ---------- Controller ----------

type webhookController struct {
	webhookSecret string
}

// NewWebhookController creates a new webhook controller.
// webhookSecret is the Resend/Svix signing secret (e.g. "whsec_...").
func NewWebhookController(webhookSecret string) *webhookController {
	return &webhookController{webhookSecret: webhookSecret}
}

// RegisterRoutes registers the webhook route.
func (c *webhookController) RegisterRoutes(app *fiber.App) {
	app.Post("/v1/webhooks/resend", c.HandleInboundWebhook)
}

// HandleInboundWebhook verifies the Svix signature and processes Resend webhook events.
func (c *webhookController) HandleInboundWebhook(ctx fiber.Ctx) error {
	payload := ctx.Body()

	// Build an http.Header from the Svix headers that Resend sends.
	headers := http.Header{}
	for _, key := range []string{"svix-id", "svix-timestamp", "svix-signature"} {
		if v := ctx.Get(key); v != "" {
			headers.Set(key, v)
		}
	}

	// Verify the webhook signature.
	wh, err := svix.NewWebhook(c.webhookSecret)
	if err != nil {
		log.Errorf("invalid webhook secret: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "webhook verification misconfigured")
	}
	if err := wh.Verify(payload, headers); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid webhook signature")
	}

	// Signature valid — parse the event.
	webhook, err := ParseWebhookEvent(payload)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	switch webhook.Type {
	case "email.received":
		c.handleEmailReceived(webhook.Data.(EmailReceived))
	case "email.bounced":
		c.handleEmailBounced(webhook.Data.(EmailBounced))
	default:
		log.Infof("Received unhandled webhook type: %s", webhook.Type)
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (c *webhookController) handleEmailReceived(data EmailReceived) {
	log.Infof("Received email webhook: %s", data.EmailID)

}

func (c *webhookController) handleEmailBounced(data EmailBounced) {
	log.Infof("Received bounce webhook: %s", data.EmailID)

}
