package worker

import (
	"context"
	"log"
	"time"

	"SpaceBookProject/internal/domain"
)

type BookingEventWorker struct {
	Events <-chan domain.BookingEvent
}

func NewBookingEventWorker(events <-chan domain.BookingEvent) *BookingEventWorker {
	return &BookingEventWorker{Events: events}
}

func (w *BookingEventWorker) Run(ctx context.Context) {
	log.Println("[worker] booking event worker started")
	defer log.Println("[worker] booking event worker stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-w.Events:
			log.Printf(
				"[worker] event=%s booking_id=%d space_id=%d tenant_id=%d at=%s\n",
				evt.Type, evt.BookingID, evt.SpaceID, evt.TenantID,
				evt.At.Format(time.RFC3339),
			)
			// TODO: here insert into DB, send email, etc.
		}
	}
}
