package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"hotel-test/internal/domain"
	"hotel-test/internal/service"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.Username]; exists {
		return domain.ErrAlreadyExists
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[username]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepo()
	jwtSecret := "test-secret-key-12345678901234567890"
	authSvc := service.NewAuthService(repo, jwtSecret, time.Hour, 4)

	ctx := context.Background()

	regReq := domain.RegisterRequest{
		Username: "hotel_manager",
		Password: "strongpassword123",
	}

	authResp, err := authSvc.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	if authResp.Token == "" {
		t.Errorf("expected non-empty token")
	}
	if authResp.User.Username != regReq.Username {
		t.Errorf("expected username %s, got %s", regReq.Username, authResp.User.Username)
	}

	savedUser, _ := repo.GetByUsername(ctx, "hotel_manager")
	if savedUser.PasswordHash == regReq.Password {
		t.Fatalf("security violation: password was stored as plaintext!")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(savedUser.PasswordHash), []byte(regReq.Password)); err != nil {
		t.Errorf("password hash does not match original password: %v", err)
	}

	_, err = authSvc.Register(ctx, regReq)
	if err == nil {
		t.Errorf("expected error registering duplicate username, got nil")
	}

	loginResp, err := authSvc.Login(ctx, domain.LoginRequest{
		Username: "hotel_manager",
		Password: "strongpassword123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.Token == "" {
		t.Errorf("expected token on successful login")
	}

	_, err = authSvc.Login(ctx, domain.LoginRequest{
		Username: "hotel_manager",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Errorf("expected unauthorized error for wrong password, got nil")
	}

	claims, err := authSvc.ValidateToken(loginResp.Token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.Username != "hotel_manager" {
		t.Errorf("expected username in token claims to be 'hotel_manager', got %s", claims.Username)
	}

	_, err = authSvc.ValidateToken("invalid.token.string")
	if err == nil {
		t.Errorf("expected error on invalid token, got nil")
	}
}
