package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const DateFormat = "2006-01-02"

type Date struct {
	time.Time
}

func NewDate(t time.Time) Date {
	y, m, d := t.Date()
	return Date{Time: time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(DateFormat, strings.TrimSpace(s))
	if err != nil {
		return Date{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return Date{Time: t}, nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Format(DateFormat))
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if strings.TrimSpace(s) == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Format(DateFormat), nil
}

func (d *Date) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*d = NewDate(v)
		return nil
	case string:
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case []byte:
		parsed, err := ParseDate(string(v))
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan %T into domain.Date", value)
	}
}

type GuestRecord struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	RoomNumber   string    `json:"room_number"`
	CheckInDate  Date      `json:"check_in_date"`
	CheckOutDate Date      `json:"check_out_date"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

var phoneRegex = regexp.MustCompile(`^\+?[0-9\s\-\(\)]{6,32}$`)

type CreateGuestRequest struct {
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	RoomNumber   string `json:"room_number"`
	CheckInDate  Date   `json:"check_in_date"`
	CheckOutDate Date   `json:"check_out_date"`
	Note         string `json:"note"`
}

func (r *CreateGuestRequest) Validate() error {
	r.FullName = strings.TrimSpace(r.FullName)
	if len(r.FullName) < 2 || len(r.FullName) > 255 {
		return fmt.Errorf("%w: full_name must be between 2 and 255 characters", ErrInvalidInput)
	}

	r.Phone = strings.TrimSpace(r.Phone)
	if !phoneRegex.MatchString(r.Phone) {
		return fmt.Errorf("%w: phone must be a valid phone number (6-32 chars)", ErrInvalidInput)
	}

	r.RoomNumber = strings.TrimSpace(r.RoomNumber)
	if r.RoomNumber == "" || len(r.RoomNumber) > 32 {
		return fmt.Errorf("%w: room_number is required (up to 32 chars)", ErrInvalidInput)
	}

	if r.CheckInDate.IsZero() {
		return fmt.Errorf("%w: check_in_date is required", ErrInvalidInput)
	}
	if r.CheckOutDate.IsZero() {
		return fmt.Errorf("%w: check_out_date is required", ErrInvalidInput)
	}

	if !r.CheckOutDate.After(r.CheckInDate.Time) {
		return fmt.Errorf("%w: check_out_date (%s) must be after check_in_date (%s)",
			ErrDateConflict, r.CheckOutDate.Format(DateFormat), r.CheckInDate.Format(DateFormat))
	}

	if len(r.Note) > 2000 {
		return fmt.Errorf("%w: note cannot exceed 2000 characters", ErrInvalidInput)
	}

	return nil
}

type UpdateGuestRequest struct {
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	RoomNumber   string `json:"room_number"`
	CheckInDate  Date   `json:"check_in_date"`
	CheckOutDate Date   `json:"check_out_date"`
	Note         string `json:"note"`
}

func (r *UpdateGuestRequest) Validate() error {
	req := CreateGuestRequest{
		FullName:     r.FullName,
		Phone:        r.Phone,
		RoomNumber:   r.RoomNumber,
		CheckInDate:  r.CheckInDate,
		CheckOutDate: r.CheckOutDate,
		Note:         r.Note,
	}
	err := req.Validate()
	if err != nil {
		return err
	}
	r.FullName = req.FullName
	r.Phone = req.Phone
	r.RoomNumber = req.RoomNumber
	return nil
}

type GuestFilter struct {
	RoomNumber string
	Phone      string
	Limit      int
	Offset     int
}
