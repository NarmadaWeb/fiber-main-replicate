package handlers

import (
	"errors" // Added for error checking
	"fmt"
	"log"
	"time"

	"github.com/NarmadaWeb/fiber-replicate-example/internal/config"
	"github.com/NarmadaWeb/fiber-replicate-example/internal/stores" // Added for ItemStore
	"github.com/gofiber/fiber/v2"
)

// ReplicaItemHandler holds dependencies for replica item handlers.
type ReplicaItemHandler struct {
	Store stores.ItemStore
}

// NewReplicaItemHandler creates a new ReplicaItemHandler instance.
func NewReplicaItemHandler(s stores.ItemStore) *ReplicaItemHandler {
	if s == nil {
		log.Fatal("FATAL: ReplicaItemHandler requires a non-nil ItemStore")
	}
	return &ReplicaItemHandler{Store: s}
}

// --- Standalone Handlers (No Store Dependency) ---

func LocalDataHandlerReplica(c *fiber.Ctx) error {
	log.Println("INFO: Replica Server '/local-data' endpoint hit")
	data := fiber.Map{
		"id":          "replica-local-data-001",
		"source":      "Replica Server (Local Cache/Static)",
		"description": "Ini adalah data statis yang ada di Replica Server.",
		"retrievedAt": time.Now().Format(time.RFC3339),
		"configPort":  config.AppConfig.ReplicaServer.Port,
	}
	return c.JSON(data)
}

func FetchFromMainHandler(c *fiber.Ctx) error {
	agent := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(agent)

	// Construct URL carefully, ensure mainServerAddress includes scheme (http/https)
	mainDataURL := config.AppConfig.ReplicaServer.MainServerAddress + "/api/v1/items" // Assuming /items is the target
	timeout := config.GetReplicaFetchTimeout()

	log.Printf("INFO: Replica '/fetch-from-main' endpoint hit. Fetching from %s with timeout %s", mainDataURL, timeout)

	agent.Timeout(timeout)
	req := agent.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI(mainDataURL)

	// Parsing agent request
	if err := agent.Parse(); err != nil {
		log.Printf("ERROR: Replica '/fetch-from-main' - Failed to prepare request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal proxy error preparing request"})
	}

	// Eksekusi request
	code, bodyBytes, errs := agent.Bytes()

	// Cek errors (koneksi, timeout, dll)
	if len(errs) > 0 {
		errMsg := fmt.Sprintf("Failed to connect or fetch from main server (%s)", mainDataURL)
		log.Printf("ERROR: Replica '/fetch-from-main' - %s: %v", errMsg, errs[0])
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":     errMsg,
			"targetUrl": mainDataURL,
			"details":   errs[0].Error(),
		})
	}

	// Cek status code dari Main Server
	if code != fiber.StatusOK {
		errMsg := fmt.Sprintf("Main server (%s) returned non-OK status", mainDataURL)
		log.Printf("WARN: Replica '/fetch-from-main' - %s: Status %d, Body: %s", errMsg, code, string(bodyBytes))
		// Forward status code dan body dari Main
		// Content-Type is usually handled by Fiber automatically when sending bytes,
		// or the client needs to interpret the raw bytes. Removing problematic agent.Response() call.
		// c.Set(fiber.HeaderContentType, agent.Response().Header.ContentType())
		return c.Status(code).Send(bodyBytes)
	}

	log.Printf("INFO: Replica '/fetch-from-main' - Successfully fetched data from %s, status %d", mainDataURL, code)

	// Forward response sukses
	// Content-Type is usually handled by Fiber automatically when sending bytes.
	// Removing problematic agent.Response() call.
	// c.Set(fiber.HeaderContentType, agent.Response().Header.ContentType())
	return c.Status(fiber.StatusOK).Send(bodyBytes)
}

// --- Replica Item Handlers (Using Store) ---

// GetItemsReplica handles GET /api/v1/items on the replica.
func (h *ReplicaItemHandler) GetItemsReplica(c *fiber.Ctx) error {
	log.Println("INFO: GetItemsReplica - Request received")
	items, err := h.Store.GetAllItems()
	if err != nil {
		log.Printf("ERROR: GetItemsReplica - Handler received error from store: %v", err)
		// Consider specific error handling if needed (e.g., ErrDatabaseError)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve items from replica store"})
	}

	log.Printf("INFO: GetItemsReplica - Returning %d items via handler from replica store", len(items))
	return c.Status(fiber.StatusOK).JSON(items)
}

// GetItemReplica handles GET /api/v1/items/:id on the replica.
func (h *ReplicaItemHandler) GetItemReplica(c *fiber.Ctx) error {
	itemID := c.Params("id")
	log.Printf("INFO: GetItemReplica - Request received for ID: %s", itemID)

	item, err := h.Store.GetItem(itemID)
	if err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			log.Printf("INFO: GetItemReplica - Item not found via handler: ID %s", itemID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": fmt.Sprintf("Item with ID '%s' not found on replica", itemID),
			})
		}
		log.Printf("ERROR: GetItemReplica - Handler received error from store: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve item from replica store"})
	}

	log.Printf("INFO: GetItemReplica - Item found via handler: ID %d", item.ID)
	return c.Status(fiber.StatusOK).JSON(item)
}

// --- Removed Proxy Logic ---
/*
func proxyRequest(c *fiber.Ctx, method, path string) error {
	// ... (proxy logic removed as handlers now use local store) ...
}

func GetItemsReplica(c *fiber.Ctx) error {
	return proxyRequest(c, fiber.MethodGet, "/api/v1/items")
}

func GetItemReplica(c *fiber.Ctx) error {
	itemID := c.Params("id")
	path := fmt.Sprintf("/api/v1/items/%s", itemID)
	return proxyRequest(c, fiber.MethodGet, path)
}
*/
