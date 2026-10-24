package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"hotel-test/internal/domain"
	"hotel-test/internal/handler"
	"hotel-test/internal/service"

	"github.com/google/uuid"
)

type mockUserRepo struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[username]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
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

type mockGuestRepo struct {
	mu      sync.RWMutex
	records map[uuid.UUID]domain.GuestRecord
}

func (m *mockGuestRepo) Create(ctx context.Context, guest *domain.GuestRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if guest.ID == uuid.Nil {
		guest.ID = uuid.New()
	}
	guest.CreatedAt = time.Now()
	guest.UpdatedAt = time.Now()
	m.records[guest.ID] = *guest
	return nil
}

func (m *mockGuestRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.GuestRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.records[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &r, nil
}

func (m *mockGuestRepo) List(ctx context.Context, filter domain.GuestFilter) ([]domain.GuestRecord, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []domain.GuestRecord
	for _, r := range m.records {
		list = append(list, r)
	}
	return list, len(list), nil
}

func (m *mockGuestRepo) Update(ctx context.Context, guest *domain.GuestRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.records[guest.ID]; !ok {
		return domain.ErrNotFound
	}
	guest.UpdatedAt = time.Now()
	m.records[guest.ID] = *guest
	return nil
}

func (m *mockGuestRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.records[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.records, id)
	return nil
}

func setupTestRouter() (http.Handler, *service.AuthService) {
	userRepo := &mockUserRepo{users: make(map[string]*domain.User)}
	guestRepo := &mockGuestRepo{records: make(map[uuid.UUID]domain.GuestRecord)}

	authService := service.NewAuthService(userRepo, "test-super-secret-key-1234567890123", time.Hour, 4)
	guestService := service.NewGuestService(guestRepo)

	router := handler.NewRouter(authService, guestService)
	return router, authService
}

func TestGuestEndpoints_UnauthorizedWithoutToken(t *testing.T) {
	router, _ := setupTestRouter()

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/guests"},
		{"POST", "/api/v1/guests"},
		{"GET", "/api/v1/guests/" + uuid.New().String()},
		{"PUT", "/api/v1/guests/" + uuid.New().String()},
		{"DELETE", "/api/v1/guests/" + uuid.New().String()},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("[%s %s] expected 401 Unauthorized, got %d", ep.method, ep.path, rec.Code)
		}
	}
}

func TestGuestEndpoints_FullCRUDWorkflow(t *testing.T) {
	router, _ := setupTestRouter()

	regBody, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "secretpassword",
	})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)

	if regRec.Code != http.StatusCreated {
		t.Fatalf("expected register 201 Created, got %d: %s", regRec.Code, regRec.Body.String())
	}

	var authResp domain.AuthResponse
	if err := json.Unmarshal(regRec.Body.Bytes(), &authResp); err != nil {
		t.Fatalf("failed to parse auth response: %v", err)
	}
	token := authResp.Token

	invalidPayload := `{
		"full_name": "Алексеев Алексей",
		"phone": "+79990001122",
		"room_number": "205",
		"check_in_date": "2026-10-30",
		"check_out_date": "2026-10-25"
	}`
	badReq := httptest.NewRequest("POST", "/api/v1/guests", bytes.NewReader([]byte(invalidPayload)))
	badReq.Header.Set("Authorization", "Bearer "+token)
	badReq.Header.Set("Content-Type", "application/json")
	badRec := httptest.NewRecorder()
	router.ServeHTTP(badRec, badReq)

	if badRec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for inverted dates, got %d: %s", badRec.Code, badRec.Body.String())
	}

	validPayload := `{
		"full_name": "Смирнова Анна Сергеевна",
		"phone": "+7 (912) 345-67-89",
		"room_number": "301-Deluxe",
		"check_in_date": "2026-11-01",
		"check_out_date": "2026-11-05",
		"note": "Ранний заезд, дополнительное одеяло"
	}`
	createReq := httptest.NewRequest("POST", "/api/v1/guests", bytes.NewReader([]byte(validPayload)))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", createRec.Code, createRec.Body.String())
	}

	var createdGuest domain.GuestRecord
	if err := json.Unmarshal(createRec.Body.Bytes(), &createdGuest); err != nil {
		t.Fatalf("failed to decode created guest: %v", err)
	}
	if createdGuest.ID == uuid.Nil {
		t.Errorf("expected created guest to have a valid UUID")
	}
	if createdGuest.FullName != "Смирнова Анна Сергеевна" {
		t.Errorf("expected name 'Смирнова Анна Сергеевна', got '%s'", createdGuest.FullName)
	}

	getReq := httptest.NewRequest("GET", "/api/v1/guests/"+createdGuest.ID.String(), nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", getRec.Code)
	}

	updatePayload := `{
		"full_name": "Смирнова Анна Сергеевна",
		"phone": "+7 (912) 345-67-89",
		"room_number": "302-Suite",
		"check_in_date": "2026-11-01",
		"check_out_date": "2026-11-07",
		"note": "Повышен класс номера до Suite"
	}`
	updReq := httptest.NewRequest("PUT", "/api/v1/guests/"+createdGuest.ID.String(), bytes.NewReader([]byte(updatePayload)))
	updReq.Header.Set("Authorization", "Bearer "+token)
	updReq.Header.Set("Content-Type", "application/json")
	updRec := httptest.NewRecorder()
	router.ServeHTTP(updRec, updReq)

	if updRec.Code != http.StatusOK {
		t.Errorf("expected 200 OK on update, got %d: %s", updRec.Code, updRec.Body.String())
	}

	delReq := httptest.NewRequest("DELETE", "/api/v1/guests/"+createdGuest.ID.String(), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content on delete, got %d", delRec.Code)
	}

	getDeletedReq := httptest.NewRequest("GET", "/api/v1/guests/"+createdGuest.ID.String(), nil)
	getDeletedReq.Header.Set("Authorization", "Bearer "+token)
	getDeletedRec := httptest.NewRecorder()
	router.ServeHTTP(getDeletedRec, getDeletedReq)

	if getDeletedRec.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for deleted guest, got %d", getDeletedRec.Code)
	}
}
