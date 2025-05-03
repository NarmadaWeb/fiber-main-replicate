package handlers

import (
	"errors"
	"fmt"
	"log"

	"github.com/NarmadaWeb/fiber-replicate-example/internal/models"
	"github.com/NarmadaWeb/fiber-replicate-example/internal/stores"
	"github.com/gofiber/fiber/v2"
)

// ItemHandler menampung dependensi untuk handler item (yaitu store).
type ItemHandler struct {
	Store stores.ItemStore // Inject store interface
}

// NewItemHandler membuat instance ItemHandler baru.
func NewItemHandler(s stores.ItemStore) *ItemHandler {
	if s == nil {
		log.Fatal("FATAL: ItemHandler requires a non-nil ItemStore")
	}
	return &ItemHandler{Store: s}
}

// CreateItem menangani POST /api/v1/items
func (h *ItemHandler) CreateItem(c *fiber.Ctx) error {
	req := new(models.CreateItemRequest)

	// Parse request body
	if err := c.BodyParser(req); err != nil {
		log.Printf("WARN: CreateItem - Bad request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body format",
			"details": err.Error(),
		})
	}

	if req.Name == "" || len(req.Name) < 3 || len(req.Name) > 255 {
		log.Printf("WARN: CreateItem - Validation failed: Name '%s'", req.Name)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validation failed", "details": "Name is required, min 3, max 255 characters"})
	}
	if req.Value < 0 {
		log.Printf("WARN: CreateItem - Validation failed: Value '%d'", req.Value)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validation failed", "details": "Value cannot be negative"})
	}


	log.Printf("INFO: CreateItem - Received valid request: %+v", req)

	newItem, err := h.Store.CreateItem(req)
	if err != nil {
		log.Printf("ERROR: CreateItem - Handler received error from store: %v", err)
		if errors.Is(err, stores.ErrDatabaseError) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error during creation"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create item"})
	}

	log.Printf("INFO: CreateItem - Item created successfully via handler: ID %d", newItem.ID)
	return c.Status(fiber.StatusCreated).JSON(newItem)
}

func (h *ItemHandler) GetItems(c *fiber.Ctx) error {
	log.Println("INFO: GetItems - Request received")
	items, err := h.Store.GetAllItems()
	if err != nil {
		log.Printf("ERROR: GetItems - Handler received error from store: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve items"})
	}

	log.Printf("INFO: GetItems - Returning %d items via handler", len(items))
	return c.Status(fiber.StatusOK).JSON(items)
}

func (h *ItemHandler) GetItem(c *fiber.Ctx) error {
	itemID := c.Params("id")
	log.Printf("INFO: GetItem - Request received for ID: %s", itemID)

	item, err := h.Store.GetItem(itemID)
	if err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			// Store sudah logging
			log.Printf("INFO: GetItem - Item not found via handler: ID %s", itemID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": fmt.Sprintf("Item with ID '%s' not found", itemID),
			})
		}
		log.Printf("ERROR: GetItem - Handler received error from store: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve item"})
	}

	log.Printf("INFO: GetItem - Item found via handler: ID %d", item.ID)
	return c.Status(fiber.StatusOK).JSON(item)
}

func (h *ItemHandler) UpdateItem(c *fiber.Ctx) error {
	itemID := c.Params("id")
	req := new(models.UpdateItemRequest)

	if err := c.BodyParser(req); err != nil {
		log.Printf("WARN: UpdateItem - Bad request body for ID %s: %v", itemID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body format",
			"details": err.Error(),
		})
	}

	if req.Name != nil && (len(*req.Name) < 3 || len(*req.Name) > 255) {
		log.Printf("WARN: UpdateItem - Validation failed for ID %s: Name '%s'", itemID, *req.Name)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validation failed", "details": "Name must be between 3 and 255 characters if provided"})
	}
	if req.Value != nil && *req.Value < 0 {
		log.Printf("WARN: UpdateItem - Validation failed for ID %s: Value '%d'", itemID, *req.Value)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validation failed", "details": "Value cannot be negative if provided"})
	}

	log.Printf("INFO: UpdateItem - Received valid update request for ID %s: %+v", itemID, req)

	updatedItem, err := h.Store.UpdateItem(itemID, req)
	if err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			log.Printf("INFO: UpdateItem - Item not found via handler: ID %s", itemID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": fmt.Sprintf("Item with ID '%s' not found for update", itemID),
			})
		}
		if errors.Is(err, stores.ErrInvalidUpdate) {
			log.Printf("INFO: UpdateItem - Invalid update via handler (no fields): ID %s", itemID)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid Update",
				"message": "No valid fields provided for update.",
			})
		}
		log.Printf("ERROR: UpdateItem - Handler received error from store: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update item"})
	}

	log.Printf("INFO: UpdateItem - Item updated successfully via handler: ID %d", updatedItem.ID)
	return c.Status(fiber.StatusOK).JSON(updatedItem)
}

func (h *ItemHandler) DeleteItem(c *fiber.Ctx) error {
	itemID := c.Params("id")
	log.Printf("INFO: DeleteItem - Received delete request for ID: %s", itemID)

	err := h.Store.DeleteItem(itemID)
	if err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			log.Printf("INFO: DeleteItem - Item not found via handler: ID %s", itemID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": fmt.Sprintf("Item with ID '%s' not found for deletion", itemID),
			})
		}
		log.Printf("ERROR: DeleteItem - Handler received error from store: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete item"})
	}

	log.Printf("INFO: DeleteItem - Item deleted successfully via handler: ID %s", itemID)
	return c.SendStatus(fiber.StatusNoContent)
}
