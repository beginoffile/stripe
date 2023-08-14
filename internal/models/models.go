package models

import (
	"database/sql"
	"time"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

// NewModels returns a model type with database connection pool
func NewModel(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

// Widget is the type for all widgets
type Widget struct {
	ID             int       `json: "id"`
	Name           string    `json: "name"`
	Description    string    `json: "description"`
	InventoryLevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	CreateAt       time.Time `json:"_"`
	UpdatedAt      time.Time `json:"_"`
}
