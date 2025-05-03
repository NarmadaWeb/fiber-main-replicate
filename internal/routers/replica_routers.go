package routers

import (
	"fmt"
	"log"

	"github.com/NarmadaWeb/fiber-replicate-example/internal/config" // Added import
	"github.com/NarmadaWeb/fiber-replicate-example/internal/handlers"
	"github.com/NarmadaWeb/fiber-replicate-example/internal/stores" // Added import
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupReplicaRoutes configures routes for the replica server.
// It now accepts an ItemStore to provide DB access to handlers.
func SetupReplicaRoutes(app *fiber.App, itemStore stores.ItemStore) { // Added itemStore parameter
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] [REPLICA] ${status} | ${latency} | ${ip} | ${method} ${path}\n",
	}))

	// --- Instantiate Handlers ---
	// Common handlers (don't need store)
	// Replica item handlers (need store)
	replicaItemHandler := handlers.NewReplicaItemHandler(itemStore)

	// --- Rute API v1 ---
	api := app.Group("/api/v1")

	// Routes without store dependency
	api.Get("/", handlers.RootHandlerReplica) // Assuming this exists in common_handlers or replica_handlers
	api.Get("/health", handlers.HealthCheck)   // Assuming this exists in common_handlers
	api.Get("/local-data", handlers.LocalDataHandlerReplica)
	api.Get("/fetch-from-main", handlers.FetchFromMainHandler)

	// Routes requiring the item store (using ReplicaItemHandler methods)
	itemsGroup := api.Group("/items")
	itemsGroup.Get("/", replicaItemHandler.GetItemsReplica)   // Use method from handler instance
	itemsGroup.Get("/:id", replicaItemHandler.GetItemReplica) // Use method from handler instance

	// Fallback for undefined routes on replica (prevents accidental writes)
	app.Use(func(c *fiber.Ctx) error {
		log.Printf("WARN: Route not found on Replica Server: %s %s", c.Method(), c.Path())
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": fmt.Sprintf("The requested resource '%s' was not found or the method '%s' is not supported on this replica server (%s). Write operations (POST, PUT, DELETE) should be directed to the main server.", c.Path(), c.Method(), config.AppConfig.ReplicaServer.AppName), // Used imported config
		})
	})
}
