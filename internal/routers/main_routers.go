package routers

import (
	"fmt"
	"log"

	"github.com/NarmadaWeb/fiber-replicate-example/internal/config"
	"github.com/NarmadaWeb/fiber-replicate-example/internal/handlers"
	"github.com/NarmadaWeb/fiber-replicate-example/internal/stores"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)


func SetupMainRoutes(app *fiber.App, itemStore stores.ItemStore) {
	// Middleware Global
	app.Use(recover.New()) // Recover dari panic
	app.Use(logger.New(logger.Config{
		// Format log: [TIMESTAMP] [MAIN] STATUS | LATENCY | CLIENT_IP | METHOD PATH
		Format: "[${time}] [MAIN] ${status} | ${latency} | ${ip} | ${method} ${path}\n",
	}))

	itemHandler := handlers.NewItemHandler(itemStore)

	// --- Rute API v1 ---
	api := app.Group("/api/v1")

	// Rute Umum
	api.Get("/", handlers.RootHandlerMain)
	api.Get("/health", handlers.HealthCheck)

	itemsGroup := api.Group("/items")
	itemsGroup.Post("/", itemHandler.CreateItem)
	itemsGroup.Get("/", itemHandler.GetItems)
	itemsGroup.Get("/:id", itemHandler.GetItem)
	itemsGroup.Put("/:id", itemHandler.UpdateItem)
	itemsGroup.Delete("/:id", itemHandler.DeleteItem)

	api.Get("/data", handlers.LocalDataHandlerReplica)


	app.Use(func(c *fiber.Ctx) error {
		log.Printf("WARN: Route not found on Main Server: %s %s", c.Method(), c.Path())
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": fmt.Sprintf("The requested resource '%s' was not found on this server (%s).", c.Path(), config.AppConfig.MainServer.AppName),
		})
	})
}
