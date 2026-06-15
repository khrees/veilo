package controllers

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/khrees/veilo/services"
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
	EmailID     string       `json:"email_id"`
	CreatedAt   time.Time    `json:"created_at"`
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Bcc         []string     `json:"bcc"`
	Cc          []string     `json:"cc"`
	MessageID   string       `json:"message_id"`
	Subject     string       `json:"subject"`
	Attachments []Attachment `json:"attachments"`
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
	webhookSvc    services.WebhookService
}

// NewWebhookController creates a new webhook controller.
func NewWebhookController(deps RouteDeps) *webhookController {
	return &webhookController{
		webhookSecret: deps.WebhookSecret,
		webhookSvc:    deps.WebhookSvc,
	}
}

// RegisterRoutes registers the webhook routes.
func (c *webhookController) RegisterRoutes(app *fiber.App) {
	app.Post("/", c.HandleInboundWebhook)
	app.Post("/webhook/inbound", c.HandleInboundWebhook)
	// needed for testing with smee.io
	app.Post("/:channel/webhook/inbound", c.HandleInboundWebhook)
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

	// Verify the webhook signature first.
	wh, err := svix.NewWebhook(c.webhookSecret)
	if err != nil {
		log.Errorf("invalid webhook secret: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "webhook verification misconfigured")
	}
	if err := wh.Verify(payload, headers); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid webhook signature")
	}

	// Parse the event.
	webhook, err := ParseWebhookEvent(payload)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid webhook payload")
	}

	switch webhook.Type {
	case "email.received":
		emailData, ok := webhook.Data.(EmailReceived)
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "invalid email.received data")
		}

		// Validate from address format
		if _, err := mail.ParseAddress(strings.TrimSpace(emailData.From)); err != nil {
			log.Warnf("webhook: invalid from address: %s", emailData.From)
			return fiber.NewError(fiber.StatusBadRequest, "invalid 'from' address")
		}

		// Validate to addresses list
		if len(emailData.To) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "no 'to' addresses provided")
		}
		for _, email := range emailData.To {
			if _, err := mail.ParseAddress(strings.TrimSpace(email)); err != nil {
				log.Warnf("webhook: invalid to address: %s", email)
				return fiber.NewError(fiber.StatusBadRequest, "invalid 'to' address")
			}
		}

		// Validate subject length
		if len(emailData.Subject) > 256 {
			log.Warnf("webhook: subject too long: %d characters", len(emailData.Subject))
			return fiber.NewError(fiber.StatusBadRequest, "subject exceeds 256 characters")
		}

		input := services.EmailReceivedInput{
			EmailID:   emailData.EmailID,
			From:      emailData.From,
			To:        emailData.To,
			MessageID: emailData.MessageID,
			Subject:   emailData.Subject,
		}

		err = c.webhookSvc.ProcessEmailReceived(ctx.Context(), input)
		if err != nil {
			log.Errorf("failed to process inbound email webhook: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to process inbound email")
		}

		return ctx.SendStatus(fiber.StatusOK)

	case "email.bounced":
		emailBounced, ok := webhook.Data.(EmailBounced)
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "invalid email.bounced data")
		}
		log.Infof("Received bounce webhook: %s", emailBounced.EmailID)
		return ctx.SendStatus(fiber.StatusOK)

	default:
		log.Infof("Received unhandled webhook type: %s", webhook.Type)
		return ctx.SendStatus(fiber.StatusOK)
	}
}
