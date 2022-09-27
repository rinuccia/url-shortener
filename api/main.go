package main

import (
	"github.com/gofiber/fiber"
	"github.com/joho/godotenv"
	"github.com/rinuccia/url-shortener/routes"
	"github.com/sirupsen/logrus"
	"os"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/urls", routes.ShortenURL)
}

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
	}

	app := fiber.New()

	setupRoutes(app)

	if err := app.Listen(os.Getenv("APP_PORT")); err != nil {
		logrus.Fatalf("error occurred while running http server: %s", err.Error())
	}
}
