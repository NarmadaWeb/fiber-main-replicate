package stores

import "errors"

var (
	ErrNotFound      = errors.New("item not found")
	ErrDuplicateID   = errors.New("item with this ID already exists")
	ErrInvalidUpdate = errors.New("invalid update data")
	ErrDatabaseError = errors.New("database operation failed")
)
