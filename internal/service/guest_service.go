package service

import (
	"context"
	"fmt"

	"hotel-test/internal/domain"
	"hotel-test/internal/repository"

	"github.com/google/uuid"
)

type GuestService struct {
	repo repository.GuestRepository
}

func NewGuestService(repo repository.GuestRepository) *GuestService {
	return &GuestService{repo: repo}
}

func (s *GuestService) Create(ctx context.Context, req domain.CreateGuestRequest) (*domain.GuestRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	record := &domain.GuestRecord{
		FullName:     req.FullName,
		Phone:        req.Phone,
		RoomNumber:   req.RoomNumber,
		CheckInDate:  req.CheckInDate,
		CheckOutDate: req.CheckOutDate,
		Note:         req.Note,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create guest record: %w", err)
	}

	return record, nil
}

func (s *GuestService) GetByID(ctx context.Context, id uuid.UUID) (*domain.GuestRecord, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GuestService) List(ctx context.Context, filter domain.GuestFilter) ([]domain.GuestRecord, int, error) {
	return s.repo.List(ctx, filter)
}

func (s *GuestService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateGuestRequest) (*domain.GuestRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.FullName = req.FullName
	existing.Phone = req.Phone
	existing.RoomNumber = req.RoomNumber
	existing.CheckInDate = req.CheckInDate
	existing.CheckOutDate = req.CheckOutDate
	existing.Note = req.Note

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update guest record: %w", err)
	}

	return existing, nil
}

func (s *GuestService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
