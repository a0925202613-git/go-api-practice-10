package models

import "time"

type Event struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"        binding:"required,max=200"`
	Venue      string    `json:"venue"       binding:"required,max=200"`
	Price      int       `json:"price"       binding:"required,gte=0"`
	TotalStock int       `json:"total_stock" binding:"required,gte=1"`
	Stock      int       `json:"stock"`
	Available  bool      `json:"available"`
	EventDate  time.Time `json:"event_date"  binding:"required"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
