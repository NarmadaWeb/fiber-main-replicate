package stores

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/NarmadaWeb/fiber-replicate-example/internal/models"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(database *gorm.DB) *GormStore {
	if database == nil {
		log.Fatal("FATAL: GormStore requires a non-nil database connection")
	}
	return &GormStore{db: database}
}

func (s *GormStore) CreateItem(itemData *models.CreateItemRequest) (*models.Items, error) {
	newItem := itemData.ToModel()

	result := s.db.Create(newItem)
	if result.Error != nil {
		log.Printf("ERROR: GormStore.CreateItem - DB Error: %v", result.Error)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
        log.Printf("WARN: GormStore.CreateItem - No rows affected during create")
        return nil, fmt.Errorf("%w: no rows affected during create", ErrDatabaseError)
    }


	log.Printf("INFO: GormStore.CreateItem - Item created with ID: %d", newItem.ID)
	return newItem, nil
}

func (s *GormStore) GetItem(idStr string) (*models.Items, error) {
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Printf("WARN: GormStore.GetItem - Invalid ID format '%s': %v", idStr, err)
		return nil, fmt.Errorf("%w: invalid id format %s", ErrNotFound, idStr)
	}


	var item models.Items
	// Cari berdasarkan primary key (ID)
	result := s.db.First(&item, uint(id)) // GORM First mencari by primary key

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("INFO: GormStore.GetItem - Item not found: ID %d", id)
			return nil, fmt.Errorf("%w: id %d", ErrNotFound, id) // Gunakan error custom
		}
		log.Printf("ERROR: GormStore.GetItem - DB Error for ID %d: %v", id, result.Error)
		return nil, fmt.Errorf("%w: failed to retrieve item %d: %v", ErrDatabaseError, id, result.Error)
	}

	return &item, nil
}

func (s *GormStore) GetAllItems() ([]*models.Items, error) {
	var items []*models.Items
	// Urutkan berdasarkan ID descending sebagai contoh
	result := s.db.Order("id desc").Find(&items)
	if result.Error != nil {
		log.Printf("ERROR: GormStore.GetAllItems - DB Error: %v", result.Error)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}
	return items, nil
}

func (s *GormStore) UpdateItem(idStr string, updateData *models.UpdateItemRequest) (*models.Items, error) {
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Printf("WARN: GormStore.UpdateItem - Invalid ID format '%s': %v", idStr, err)
		return nil, fmt.Errorf("%w: invalid id format %s", ErrNotFound, idStr)
	}

	updates := updateData.ToUpdateMap()
	if len(updates) == 0 {
		log.Printf("WARN: GormStore.UpdateItem - No fields provided for update for ID %d", id)
		return nil, fmt.Errorf("%w: no fields provided for update", ErrInvalidUpdate)
	}

	var itemToUpdate models.Items
	itemToUpdate.ID = uint(id)

	result := s.db.Model(&itemToUpdate).Updates(updates)

	if result.Error != nil {
		log.Printf("ERROR: GormStore.UpdateItem - DB Error for ID %d: %v", id, result.Error)
		return nil, fmt.Errorf("%w: failed to update item %d: %v", ErrDatabaseError, id, result.Error)
	}

	if result.RowsAffected == 0 {
		_, checkErr := s.GetItem(idStr)
		if checkErr != nil && errors.Is(checkErr, ErrNotFound) {
			 log.Printf("INFO: GormStore.UpdateItem - Item not found for update: ID %d", id)
			 return nil, fmt.Errorf("%w: id %d", ErrNotFound, id)
		}
		log.Printf("INFO: GormStore.UpdateItem - No changes detected for item ID %d", id)
		 return s.GetItem(idStr)
	}


	log.Printf("INFO: GormStore.UpdateItem - Item updated successfully: ID %d", id)


	updatedItem, err := s.GetItem(idStr)
	if err != nil {
		log.Printf("ERROR: GormStore.UpdateItem - Failed to fetch updated item ID %d: %v", id, err)
		return nil, fmt.Errorf("%w: failed to fetch item %d after update: %v", ErrDatabaseError, id, err)
	}

	return updatedItem, nil
}

func (s *GormStore) DeleteItem(idStr string) error {
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Printf("WARN: GormStore.DeleteItem - Invalid ID format '%s': %v", idStr, err)
		return fmt.Errorf("%w: invalid id format %s", ErrNotFound, idStr)
	}

	result := s.db.Delete(&models.Items{}, uint(id))

	if result.Error != nil {
		log.Printf("ERROR: GormStore.DeleteItem - DB Error for ID %d: %v", id, result.Error)
		return fmt.Errorf("%w: failed to delete item %d: %v", ErrDatabaseError, id, result.Error)
	}

	if result.RowsAffected == 0 {
		log.Printf("WARN: GormStore.DeleteItem - Item not found for deletion: ID %d", id)
		return fmt.Errorf("%w: id %d", ErrNotFound, id)
	}

	log.Printf("INFO: GormStore.DeleteItem - Item (soft) deleted successfully: ID %d", id)
	return nil
}


func RunMigrations(db *gorm.DB) error {
	log.Println("INFO: Running database auto-migrations...")

	err := db.AutoMigrate(
		&models.Items{},
	)
	if err != nil {
		log.Printf("ERROR: Failed to run database auto-migrations: %v", err)
		return err
	}
	log.Println("INFO: Database auto-migrations completed successfully.")
	return nil
}
