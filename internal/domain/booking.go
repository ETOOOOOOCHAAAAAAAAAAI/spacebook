package domain

import "time"

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusApproved  BookingStatus = "approved"
	BookingStatusRejected  BookingStatus = "rejected"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID       int           `json:"id" db:"id"`
	SpaceID  int           `json:"space_id" db:"space_id"`
	TenantID int           `json:"tenant_id" db:"tenant_id"`
	Status   BookingStatus `json:"status" db:"status"`

	DateFrom  time.Time `json:"date_from" db:"date_from"`
	DateTo    time.Time `json:"date_to" db:"date_to"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
type CreateBookingRequest struct {
	SpaceID  int    `json:"space_id" binding:"required"`
	DateFrom string `json:"date_from" binding:"required"`
	DateTo   string `json:"date_to" binding:"required"`
}

type UpdateBookingStatusRequest struct {
	Status BookingStatus `json:"status" binding:"required,oneof=pending approved rejected cancelled"`
}
