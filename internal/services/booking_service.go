package services

import (
	"errors"
	"time"

	"SpaceBookProject/internal/domain"
	"SpaceBookProject/internal/repository"
)

var (
	ErrForbiddenAction      = errors.New("forbidden action")
	ErrInvalidBookingStatus = errors.New("invalid booking status for this operation")
)

type BookingService struct {
	bookings *repository.BookingRepository
}

func NewBookingService(bookings *repository.BookingRepository) *BookingService {
	return &BookingService{bookings: bookings}
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

	// Optional: check overlaps only for already approved bookings.
	hasOverlap, err := s.bookings.ExistsOverlapApproved(req.SpaceID, from, to)
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
	return b, nil
}

func (s *BookingService) ListMyBookings(tenantID int) ([]domain.Booking, error) {
	return s.bookings.ListByTenant(tenantID)
}

func (s *BookingService) ListOwnerBookings(ownerID int) ([]domain.Booking, error) {
	return s.bookings.ListByOwner(ownerID)
}

func (s *BookingService) CancelBooking(tenantID, bookingID int) error {
	b, err := s.bookings.GetByID(bookingID)
	if err != nil {
		return err
	}

	if b.TenantID != tenantID {
		return ErrForbiddenAction
	}

	if b.Status != domain.BookingStatusPending && b.Status != domain.BookingStatusApproved {
		return ErrInvalidBookingStatus
	}

	return s.bookings.UpdateStatus(bookingID, domain.BookingStatusCancelled)
}

func (s *BookingService) ApproveBooking(ownerID, bookingID int) error {
	b, err := s.bookings.GetByID(bookingID)
	if err != nil {
		return err
	}

	ownerBookings, err := s.bookings.ListByOwner(ownerID)
	if err != nil {
		return err
	}
	isOwner := false
	for _, ob := range ownerBookings {
		if ob.ID == b.ID {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return ErrForbiddenAction
	}

	if b.Status != domain.BookingStatusPending {
		return ErrInvalidBookingStatus
	}

	hasOverlap, err := s.bookings.ExistsOverlapApproved(b.SpaceID, b.DateFrom, b.DateTo)
	if err != nil {
		return err
	}
	if hasOverlap {
		return errors.New("space is already approved for these dates")
	}

	return s.bookings.UpdateStatus(bookingID, domain.BookingStatusApproved)
}

func (s *BookingService) RejectBooking(ownerID, bookingID int) error {
	b, err := s.bookings.GetByID(bookingID)
	if err != nil {
		return err
	}

	ownerBookings, err := s.bookings.ListByOwner(ownerID)
	if err != nil {
		return err
	}
	isOwner := false
	for _, ob := range ownerBookings {
		if ob.ID == b.ID {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return ErrForbiddenAction
	}

	if b.Status != domain.BookingStatusPending {
		return ErrInvalidBookingStatus
	}

	return s.bookings.UpdateStatus(bookingID, domain.BookingStatusRejected)
}
