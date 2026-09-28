// Package fixtures provides shared test models for gormz tests.
package fixtures

// User is a basic test model.
type User struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"size:255;not null"`
	Email  string `gorm:"uniqueIndex;size:255"`
	Age    int    `gorm:"default:0"`
	Active bool   `gorm:"default:false"`
}

// Order is a test model for relations.
type Order struct {
	ID     uint    `gorm:"primaryKey"`
	UserID uint    `gorm:"index"`
	Total  float64 `gorm:"type:decimal(10,2)"`
	Status string  `gorm:"size:50"`
}

// Product is an additional test model.
type Product struct {
	ID    uint    `gorm:"primaryKey"`
	Name  string  `gorm:"size:255"`
	Price float64 `gorm:"type:decimal(10,2)"`
	Stock int
}
