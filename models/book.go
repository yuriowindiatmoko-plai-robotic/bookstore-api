package models

import "time"

type Book struct {
	ID          uint      `gorm:"primaryKey"                   json:"id"`
	Title       string    `gorm:"type:varchar(255);not null"   json:"title"`
	Author      string    `gorm:"type:varchar(255);not null"   json:"author"`
	ISBN        string    `gorm:"type:varchar(57);uniqueIndex" json:"isbn"`
	Price       float64   `gorm:"type:numeric(10,2)"           json:"price"`
	Stock       int       `gorm:"default:0"                    json:"stock"`
	Description *string   `gorm:"type:text"                    json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
