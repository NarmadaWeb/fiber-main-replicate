package models

import "gorm.io/gorm"

type Items struct {
	gorm.Model

	Name        string `json:"name" gorm:"not null;index;size:255"`
	Description string `json:"description,omitempty" gorm:"type:text"`
	Value       int    `json:"value" gorm:"not null;default:0"`
}

type CreateItemRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description,omitempty"`
	Value       int    `json:"value" validate:"gte=0"`
}

type UpdateItemRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description *string `json:"description,omitempty"`
	Value       *int    `json:"value,omitempty" validate:"omitempty,gte=0"`
}

func (r *CreateItemRequest) ToModel() *Items {
	return &Items{
		Name:        r.Name,
		Description: r.Description,
		Value:       r.Value,
	}
}

func (r *UpdateItemRequest) ToUpdateMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if r.Name != nil {
		updates["name"] = *r.Name
	}
	if r.Description != nil {
		updates["description"] = *r.Description
	}
	if r.Value != nil {
		updates["value"] = *r.Value
	}
	return updates
}
