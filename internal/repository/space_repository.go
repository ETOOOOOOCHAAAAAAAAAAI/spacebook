package repository

import (
	"database/sql"
	"time"

	"SpaceBookProject/internal/domain"
)

type SpaceRepository struct {
	db *sql.DB
}

func NewSpaceRepository(db *sql.DB) *SpaceRepository {
	return &SpaceRepository{db: db}
}

func (r *SpaceRepository) List() ([]domain.Space, error) {
	rows, err := r.db.Query(`
		SELECT id, owner_id, title, description, area_m2, price, phone, created_at, updated_at
		FROM spaces
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Space
	for rows.Next() {
		var s domain.Space
		if err := rows.Scan(
			&s.ID,
			&s.OwnerID,
			&s.Title,
			&s.Description,
			&s.AreaM2,
			&s.Price,
			&s.Phone,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

func (r *SpaceRepository) Create(space *domain.Space) error {
	now := time.Now()

	query := `
		INSERT INTO spaces (owner_id, title, description, area_m2, price, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		space.OwnerID,
		space.Title,
		space.Description,
		space.AreaM2,
		space.Price,
		space.Phone,
		now,
		now,
	).Scan(&space.ID, &space.CreatedAt, &space.UpdatedAt)

	return err
}
