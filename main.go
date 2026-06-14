// Package main starts the application.
package main

import (
	"context"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
	"github.com/khrees/veilo/config"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
	"github.com/khrees/veilo/controllers"
	"github.com/khrees/veilo/services"
	"github.com/khrees/veilo/providers"
	"github.com/resend/resend-go/v3"
)

type Config struct {
	Port               string   `env:"PORT"`
	WebhookSecret      string   `env:"WEBHOOK_SECRET"`
	WebhookURL         string   `env:"WEBHOOK_URL"`
	ResendAPIKey       string   `env:"RESEND_API_KEY"`
	CloudflareAPIToken string   `env:"CLOUDFLARE_API_TOKEN"`
	ReplyTokenTTLDays  int      `env:"REPLY_TOKEN_TTL_DAYS" envDefault:"90"`
	CORSOrigins        []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"*"`
	RateLimit          int      `env:"RATE_LIMIT" envDefault:"60"`
	APIKey             string   `env:"API_KEY"`
	ViaBrandName       string   `env:"VIA_BRAND_NAME" envDefault:"Veilo"`
}

func main() {
	_ = godotenv.Load()

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatal(err)
	}

	dbCfg := &config.DBConfig{}
	if err := env.Parse(dbCfg); err != nil {
		log.Fatal(err)
	}
	log.Infof(
		"db config loaded host=%s port=%s user=%s db=%s sslmode=%s password_len=%d database_url=%t",
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.User,
		dbCfg.DBName,
		strings.TrimSpace(dbCfg.SSLMode),
		len(dbCfg.Password),
		strings.TrimSpace(dbCfg.URL) != "",
	)

	db, err := dbCfg.Connect()
	if err != nil {
		log.Fatalf("database unavailable: %v", err)
	}
	if err := db.AutoMigrate(&models.Domain{}, &models.Alias{}, &models.ReplyToken{}, &models.ForwardLog{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Repository Layer
	domainRepo := repositories.NewDomainRepository(db)
	aliasRepo := repositories.NewAliasRepository(db)
	forwardLogRepo := repositories.NewForwardLogRepository(db)
	replyTokenRepo := repositories.NewReplyTokenRepository(db)

	// Provider Layer
	var emailProv providers.EmailProvider
	var dnsProv providers.DNSProvider

	if cfg.ResendAPIKey != "" {
		resendClient := resend.NewClient(cfg.ResendAPIKey)
		emailProv = providers.NewResendEmailProvider(resendClient)
	}

	if cfg.CloudflareAPIToken != "" {
		dnsProv = providers.NewCloudflareDNSProvider(cfg.CloudflareAPIToken)
	}

	// Auto-configure Webhook on Resend if WEBHOOK_URL is provided
	if emailProv != nil && cfg.WebhookURL != "" {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			id, secret, err := emailProv.EnsureWebhook(ctx, cfg.WebhookURL)
			if err != nil {
				log.Errorf("failed to automatically configure Resend webhook: %v", err)
			} else if secret != "" {
				log.Warnf("Automatically configured new Resend webhook pointing to %s. Please copy and paste this signing secret to your .env file: WEBHOOK_SECRET=%s", cfg.WebhookURL, secret)
			} else {
				log.Infof("Resend webhook pointing to %s is already configured (ID: %s)", cfg.WebhookURL, id)
			}
		}()
	}

	// Service Layer
	domainSvc := services.NewDomainService(domainRepo, emailProv, dnsProv)
	aliasSvc := services.NewAliasService(aliasRepo)
	forwardLogSvc := services.NewForwardLogService(forwardLogRepo)
	webhookSvc := services.NewWebhookService(aliasRepo, forwardLogRepo, replyTokenRepo, emailProv, cfg.ReplyTokenTTLDays, cfg.ViaBrandName)

	// Start background domain verification worker (checks status every 30 minutes)
	domainWorker := services.NewWorker("domain-verification", 30*time.Minute, func(ctx context.Context) error {
		return domainSvc.VerifyDomains(ctx)
	})
	domainWorker.Start(context.Background())

	// Start background reply token cleanup worker (runs every 12 hours)
	tokenCleanupWorker := services.NewWorker("reply-token-cleanup", 12*time.Hour, func(ctx context.Context) error {
		return webhookSvc.CleanupExpiredTokens(ctx)
	})
	tokenCleanupWorker.Start(context.Background())

	server := NewServer(ServerConfig{
		Port:        cfg.Port,
		CORSOrigins: cfg.CORSOrigins,
		RateLimit:   cfg.RateLimit,
	}, controllers.RouteDeps{
		DomainSvc:     domainSvc,
		AliasSvc:      aliasSvc,
		ForwardLogSvc: forwardLogSvc,
		WebhookSvc:    webhookSvc,
		WebhookSecret: cfg.WebhookSecret,
		APIKey:        cfg.APIKey,
	})
	log.Fatal(server.Start())
}
