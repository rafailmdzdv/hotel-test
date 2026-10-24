package repository

import (
	"context"

	"hotel-test/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type GuestRepository interface {
	Create(ctx context.Context, guest *domain.GuestRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.GuestRecord, error)
	List(ctx context.Context, filter domain.GuestFilter) ([]domain.GuestRecord, int, error)
	Update(ctx context.Context, guest *domain.GuestRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
}
