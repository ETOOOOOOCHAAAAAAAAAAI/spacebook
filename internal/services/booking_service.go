package services

import (
	"errors"
	"time"

	"SpaceBookProject/internal/domain"
	"SpaceBookProject/internal/repository"
)

var (
	ErrForbidden          = errors.New("forbidden")
	ErrAlreadyStarted     = errors.New("booking already started")
	ErrWrongStatus        = errors.New("invalid booking status")
	ErrOverlappingBooking = errors.New("overlapping approved booking")
)

type BookingService struct {
	bookings *repository.BookingRepository
	spaces   *repository.SpaceRepository
	events   chan<- domain.BookingEvent
}

func NewBookingService(bookings *repository.BookingRepository, spaces *repository.SpaceRepository, events chan<- domain.BookingEvent) *BookingService {
	return &BookingService{
		bookings: bookings,
		spaces:   spaces,
		events:   events,
	}
}

const dateLayout = "2006-01-02"

func (s *BookingService) CreateBooking(tenantID int, req *domain.CreateBookingRequest) (*domain.Booking, error) {
	from, err := time.Parse(dateLayout, req.DateFrom)
	if err != nil {
		return nil, err
	}
	to, err := time.Parse(dateLayout, req.DateTo)
	if err != nil {
		return nil, err
	}
	if !from.Before(to) {
		return nil, errors.New("date_from must be before date_to")
	}

	hasOverlap, err := s.bookings.HasApprovedOverlap(req.SpaceID, from, to, nil)
	if err != nil {
		return nil, err
	}
	if hasOverlap {
		return nil, errors.New("space is already booked for these dates")
	}

	b := &domain.Booking{
		SpaceID:  req.SpaceID,
		TenantID: tenantID,
		Status:   domain.BookingStatusPending,
		DateFrom: from,
		DateTo:   to,
	}

	if err := s.bookings.Create(b); err != nil {
		return nil, err
	}
	if s.events != nil {
		s.events <- domain.BookingEvent{
			Type:      domain.BookingEventCreated,
			BookingID: b.ID,
			SpaceID:   b.SpaceID,
			TenantID:  b.TenantID,
			At:        time.Now(),
		}
	}

	return b, nil
}

func (s *BookingService) ListMyBookings(tenantID int) ([]domain.Booking, error) {
	return s.bookings.ListByTenant(tenantID)
}

func (s *BookingService) ListOwnerBookings(ownerID int) ([]domain.Booking, error) {
	return s.bookings.ListByOwner(ownerID)
}

func (s *BookingService) CancelBooking(id, tenantID int) error {
	b, err := s.bookings.GetByID(id)
	if err != nil {
		return err
	}

	if b.TenantID != tenantID {
		return ErrForbidden
	}
	if time.Now().After(b.DateFrom) {
		return ErrAlreadyStarted
	}
	if b.Status != domain.BookingStatusPending && b.Status != domain.BookingStatusApproved {
		return ErrWrongStatus
	}

	if err := s.bookings.UpdateStatus(id, domain.BookingStatusCancelled); err != nil {
		return err
	}
	if s.events != nil {
		s.events <- domain.BookingEvent{
			Type:      domain.BookingEventCancelled,
			BookingID: b.ID,
			SpaceID:   b.SpaceID,
			TenantID:  b.TenantID,
			At:        time.Now(),
		}
	}

	return nil
}

func (s *BookingService) ApproveBooking(id int, ownerID int) error {
	b, err := s.bookings.GetByID(id)
	if err != nil {
		return err
	}

	sp, err := s.spaces.GetByID(b.SpaceID)
	if err != nil {
		return err
	}

	if sp.OwnerID != ownerID {
		return ErrForbidden
	}

	if b.Status != domain.BookingStatusPending {
		return ErrWrongStatus
	}

	overlap, err := s.bookings.HasApprovedOverlap(b.SpaceID, b.DateFrom, b.DateTo, &b.ID)
	if err != nil {
		return err
	}
	if overlap {
		return ErrOverlappingBooking
	}

	if err := s.bookings.UpdateStatus(id, domain.BookingStatusApproved); err != nil {
		return err
	}
	if s.events != nil {
		s.events <- domain.BookingEvent{
			Type:      domain.BookingEventApproved,
			BookingID: b.ID,
			SpaceID:   b.SpaceID,
			TenantID:  b.TenantID,
			At:        time.Now(),
		}
	}

	return nil
}

func (s *BookingService) RejectBooking(id int, ownerID int) error {
	b, err := s.bookings.GetByID(id)
	if err != nil {
		return err
	}

	sp, err := s.spaces.GetByID(b.SpaceID)
	if err != nil {
		return err
	}

	if sp.OwnerID != ownerID {
		return ErrForbidden
	}

	if b.Status != domain.BookingStatusPending {
		return ErrWrongStatus
	}

	if err := s.bookings.UpdateStatus(id, domain.BookingStatusRejected); err != nil {
		return err
	}
	if s.events != nil {
		s.events <- domain.BookingEvent{
			Type:      domain.BookingEventRejected,
			BookingID: b.ID,
			SpaceID:   b.SpaceID,
			TenantID:  b.TenantID,
			At:        time.Now(),
		}
	}

	return nil
}
