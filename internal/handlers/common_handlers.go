package handlers

import (
	"time"

	"github.com/NarmadaWeb/fiber-replicate-example/internal/config"
	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "ok",
		"appName":   c.App().Config().AppName,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// RootHandlerMain handler untuk root path Main Server.
func RootHandlerMain(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Welcome to the Main Server!",
		"appName": config.AppConfig.MainServer.AppName,
		"status":  "active",
		"apiDocs": "/api/v1/docs",
	})
}

// RootHandlerReplica handler untuk root path Replica Server.
func RootHandlerReplica(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Welcome to the Replica Server!",
		"appName": config.AppConfig.ReplicaServer.AppName,
		"status":  "active",
		"mode":    "replica (read-proxy)",
		"apiDocs": "/api/v1/docs",
	})
}
