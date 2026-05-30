package main

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/khrees/cloakee/config"
	"github.com/khrees/cloakee/models"
	"github.com/khrees/cloakee/routes"
)

type Config struct {
	Port string `env:"PORT"`
}

func main() {
	app := fiber.New()

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
		log.Printf("database unavailable: %v", err)
	} else if err := db.AutoMigrate(&models.Domain{}, &models.Alias{}, &models.ReplyToken{}, &models.ForwardLog{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	routes.SetupRoutes(app)

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	log.Fatal(app.Listen(":" + cfg.Port))
}
