package main

import (
	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
	"github.com/khrees/cloakee/config"
	"github.com/khrees/cloakee/models"
	"github.com/khrees/cloakee/repositories"
	"github.com/khrees/cloakee/routes"
	"github.com/khrees/cloakee/services"
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
