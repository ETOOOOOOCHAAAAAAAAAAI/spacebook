package repository

import (
	"database/sql"
	"errors"
	"time"

	"SpaceBookProject/internal/domain"
)

var (
	ErrBookingNotFound = errors.New("booking not found")
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(b *domain.Booking) error {
	const query = `
		INSERT INTO bookings (space_id, tenant_id, date_from, date_to, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, status, created_at, updated_at;
	`

	err := r.db.QueryRow(
		query,
		b.SpaceID,
		b.TenantID,
		b.DateFrom,
		b.DateTo,
		b.Status,
	).Scan(&b.ID, &b.Status, &b.CreatedAt, &b.UpdatedAt)

	return err
}

func (r *BookingRepository) GetByID(id int) (*domain.Booking, error) {
	const query = `
		SELECT id, space_id, tenant_id, status, date_from, date_to, created_at, updated_at
		FROM bookings
		WHERE id = $1;
	`

	b := &domain.Booking{}
	err := r.db.QueryRow(query, id).Scan(
		&b.ID,
		&b.SpaceID,
		&b.TenantID,
		&b.Status,
		&b.DateFrom,
		&b.DateTo,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	return b, nil
}

func (r *BookingRepository) ListByTenant(tenantID int) ([]domain.Booking, error) {
	const query = `
		SELECT id, space_id, tenant_id, status, date_from, date_to, created_at, updated_at
		FROM bookings
		WHERE tenant_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID,
			&b.SpaceID,
			&b.TenantID,
			&b.Status,
			&b.DateFrom,
			&b.DateTo,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, nil
}

func (r *BookingRepository) ListByOwner(ownerID int) ([]domain.Booking, error) {
	const query = `
		SELECT b.id, b.space_id, b.tenant_id, b.status, b.date_from, b.date_to, b.created_at, b.updated_at
		FROM bookings b
		JOIN spaces s ON b.space_id = s.id
		WHERE s.owner_id = $1
		ORDER BY b.created_at DESC;
	`

	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID,
			&b.SpaceID,
			&b.TenantID,
			&b.Status,
			&b.DateFrom,
			&b.DateTo,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, nil
}

func (r *BookingRepository) UpdateStatus(id int, status domain.BookingStatus) error {
	const query = `
		UPDATE bookings
		SET status = $1, updated_at = NOW()
		WHERE id = $2;
	`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepository) ExistsOverlapApproved(spaceID int, from, to time.Time) (bool, error) {
	const query = `
		SELECT COUNT(1)
		FROM bookings
		WHERE space_id = $1
		  AND status = 'approved'
		  AND NOT ($3 <= date_from OR $2 >= date_to);
		-- intervals overlap if NOT (new_to <= from OR new_from >= to)
	`

	var count int
	err := r.db.QueryRow(query, spaceID, from, to).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
