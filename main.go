// Package main starts the application.
package main

import (
	"strings"

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
	WebhookSecret      string   `env:"WEBHOOK_SECRET,required"`
	ResendAPIKey       string   `env:"RESEND_API_KEY"`
	CloudflareAPIToken string   `env:"CLOUDFLARE_API_TOKEN"`
	ReplyTokenTTLDays  int      `env:"REPLY_TOKEN_TTL_DAYS" envDefault:"90"`
	CORSOrigins        []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"*"`
	RateLimit          int      `env:"RATE_LIMIT" envDefault:"60"`
	APIKey             string   `env:"API_KEY"`
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

	// Service Layer
	domainSvc := services.NewDomainService(domainRepo, emailProv, dnsProv)
	aliasSvc := services.NewAliasService(aliasRepo)
	forwardLogSvc := services.NewForwardLogService(forwardLogRepo)
	webhookSvc := services.NewWebhookService(aliasRepo, forwardLogRepo, replyTokenRepo, emailProv, cfg.ReplyTokenTTLDays)

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
