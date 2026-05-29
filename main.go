package main

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

type Config struct {
	Port string `env:"PORT"`
}

func main() {
	app := fiber.New()

	_ = godotenv.Load()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// parse with generics
	cfg, err = env.ParseAs[Config]()
	if err != nil {
		log.Fatal(err)
	}

	// Define a route for the GET method on the root path '/'
	app.Get("/", func(c fiber.Ctx) error {
		// Send a string response to the client
		return c.SendString("Hello, World 👋!")
	})

	// Start the server
	log.Fatal(app.Listen(":" + cfg.Port))
}
