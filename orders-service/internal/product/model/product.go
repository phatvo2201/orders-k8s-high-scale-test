package product

import "time"

type Product struct {
	ID        string    `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	Quantity  int       `gorm:"not null"`
	Price     float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
