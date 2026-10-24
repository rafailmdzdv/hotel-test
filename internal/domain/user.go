package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return ErrInvalidInput
	}
	if len(r.Password) < 6 || len(r.Password) > 72 {
		return ErrPasswordLength
	}
	return nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)
	if r.Username == "" || r.Password == "" {
		return ErrInvalidInput
	}
	return nil
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
