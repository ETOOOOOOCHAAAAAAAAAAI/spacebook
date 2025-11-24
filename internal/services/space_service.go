package services

import (
	"SpaceBookProject/internal/domain"
	"SpaceBookProject/internal/repository"
)

type SpaceService struct {
	repo *repository.SpaceRepository
}

func NewSpaceService(repo *repository.SpaceRepository) *SpaceService {
	return &SpaceService{repo: repo}
}

func (s *SpaceService) ListSpaces() ([]domain.Space, error) {
	return s.repo.List()
}

func (s *SpaceService) CreateSpace(ownerID int, req *domain.CreateSpaceRequest) (*domain.Space, error) {
	space := &domain.Space{
		OwnerID:     ownerID,
		Title:       req.Title,
		Description: req.Description,
		AreaM2:      req.AreaM2,
		Price:       req.Price,
		Phone:       req.Phone,
	}

	if err := s.repo.Create(space); err != nil {
		return nil, err
	}
	return space, nil
}
