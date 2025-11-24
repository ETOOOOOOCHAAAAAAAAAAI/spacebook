package domain

import "time"

type Space struct {
	ID          int       `json:"id" db:"id"`
	OwnerID     int       `json:"owner_id" db:"owner_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	AreaM2      float64   `json:"area_m2" db:"area_m2"`
	Price       int       `json:"price" db:"price"`
	Phone       string    `json:"phone" db:"phone"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateSpaceRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description" binding:"required"`
	AreaM2      float64 `json:"area_m2" binding:"required,gt=0"`
	Price       int     `json:"price" binding:"required,gt=0"`
	Phone       string  `json:"phone" binding:"required"`
}
