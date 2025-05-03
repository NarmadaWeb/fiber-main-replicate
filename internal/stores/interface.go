package stores


import "github.com/NarmadaWeb/fiber-replicate-example/internal/models"


type ItemStore interface {
	CreateItem(itemData *models.CreateItemRequest) (*models.Items, error)
	GetItem(id string) (*models.Items, error)
	GetAllItems() ([]*models.Items, error)
	UpdateItem(id string, updateData *models.UpdateItemRequest) (*models.Items, error)
	DeleteItem(id string) error
}
