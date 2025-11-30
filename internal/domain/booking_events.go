package domain

import "time"

type BookingEventType string

const (
	BookingEventCreated   BookingEventType = "created"
	BookingEventApproved  BookingEventType = "approved"
	BookingEventRejected  BookingEventType = "rejected"
	BookingEventCancelled BookingEventType = "cancelled"
)

type BookingEvent struct {
	Type      BookingEventType `json:"type"`
	BookingID int              `json:"booking_id"`
	SpaceID   int              `json:"space_id"`
	TenantID  int              `json:"tenant_id"`
	At        time.Time        `json:"at"`
}
