package main

import (
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
	"github.com/khrees/veilo/config"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
	"github.com/khrees/veilo/routes"
	"github.com/khrees/veilo/services"
)

type Config struct {
	Port string `env:"PORT"`
}

func main() {
	_ = godotenv.Load()

	var cfg Config
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

	// Service Layer
	domainSvc := services.NewDomainService(domainRepo)
	aliasSvc := services.NewAliasService(aliasRepo)
	forwardLogSvc := services.NewForwardLogService(forwardLogRepo)

	server := NewServer(cfg.Port, routes.RouteDeps{
		DomainSvc:     domainSvc,
		AliasSvc:      aliasSvc,
		ForwardLogSvc: forwardLogSvc,
	})
	log.Fatal(server.Start())
}
