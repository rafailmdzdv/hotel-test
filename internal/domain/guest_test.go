package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"hotel-test/internal/domain"
)

func TestGuestValidation(t *testing.T) {
	validCheckIn, _ := domain.ParseDate("2026-10-25")
	validCheckOut, _ := domain.ParseDate("2026-10-28")

	tests := []struct {
		name        string
		req         domain.CreateGuestRequest
		expectError bool
	}{
		{
			name: "Valid guest record",
			req: domain.CreateGuestRequest{
				FullName:     "Иванов Иван Иванович",
				Phone:        "+7 (999) 123-45-67",
				RoomNumber:   "101-A",
				CheckInDate:  validCheckIn,
				CheckOutDate: validCheckOut,
				Note:         "Просил тихий номер",
			},
			expectError: false,
		},
		{
			name: "Missing full name",
			req: domain.CreateGuestRequest{
				FullName:     "   ",
				Phone:        "+79991234567",
				RoomNumber:   "101",
				CheckInDate:  validCheckIn,
				CheckOutDate: validCheckOut,
			},
			expectError: true,
		},
		{
			name: "Invalid phone number",
			req: domain.CreateGuestRequest{
				FullName:     "Петров Петр",
				Phone:        "not-a-phone",
				RoomNumber:   "102",
				CheckInDate:  validCheckIn,
				CheckOutDate: validCheckOut,
			},
			expectError: true,
		},
		{
			name: "Empty room number",
			req: domain.CreateGuestRequest{
				FullName:     "Петров Петр",
				Phone:        "+79991234567",
				RoomNumber:   "   ",
				CheckInDate:  validCheckIn,
				CheckOutDate: validCheckOut,
			},
			expectError: true,
		},
		{
			name: "Check-out date before check-in date",
			req: domain.CreateGuestRequest{
				FullName:     "Сидоров Алексей",
				Phone:        "+79991234567",
				RoomNumber:   "103",
				CheckInDate:  validCheckOut,
				CheckOutDate: validCheckIn,
			},
			expectError: true,
		},
		{
			name: "Check-out date same as check-in date",
			req: domain.CreateGuestRequest{
				FullName:     "Сидоров Алексей",
				Phone:        "+79991234567",
				RoomNumber:   "103",
				CheckInDate:  validCheckIn,
				CheckOutDate: validCheckIn,
			},
			expectError: true,
		},
		{
			name: "Missing check-in date",
			req: domain.CreateGuestRequest{
				FullName:     "Сидоров Алексей",
				Phone:        "+79991234567",
				RoomNumber:   "103",
				CheckOutDate: validCheckOut,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestDateJSONSerialization(t *testing.T) {
	originalDate, err := domain.ParseDate("2026-11-01")
	if err != nil {
		t.Fatalf("unexpected error parsing date: %v", err)
	}

	type TestPayload struct {
		Date domain.Date `json:"date"`
	}

	payload := TestPayload{Date: originalDate}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal date: %v", err)
	}

	expectedJSON := `{"date":"2026-11-01"}`
	if string(data) != expectedJSON {
		t.Errorf("expected JSON %s, got %s", expectedJSON, string(data))
	}

	var unmarshaled TestPayload
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal date: %v", err)
	}

	if unmarshaled.Date.Format(domain.DateFormat) != "2026-11-01" {
		t.Errorf("expected 2026-11-01, got %s", unmarshaled.Date.Format(domain.DateFormat))
	}
}

func TestDateDatabaseValueScan(t *testing.T) {
	d, _ := domain.ParseDate("2026-12-15")

	val, err := d.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}
	if val != "2026-12-15" {
		t.Errorf("expected '2026-12-15', got %v", val)
	}

	var scanned domain.Date
	if err := scanned.Scan(time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("Scan(time.Time) returned error: %v", err)
	}
	if scanned.Format(domain.DateFormat) != "2026-12-15" {
		t.Errorf("expected 2026-12-15, got %s", scanned.Format(domain.DateFormat))
	}

	var scannedStr domain.Date
	if err := scannedStr.Scan("2026-12-15"); err != nil {
		t.Fatalf("Scan(string) returned error: %v", err)
	}
	if scannedStr.Format(domain.DateFormat) != "2026-12-15" {
		t.Errorf("expected 2026-12-15, got %s", scannedStr.Format(domain.DateFormat))
	}
}
