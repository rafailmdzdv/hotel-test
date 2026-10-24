package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hotel-test/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type guestRepo struct {
	pool *pgxpool.Pool
}

func NewGuestRepository(pool *pgxpool.Pool) *guestRepo {
	return &guestRepo{pool: pool}
}

func (r *guestRepo) Create(ctx context.Context, guest *domain.GuestRecord) error {
	query := `
		INSERT INTO guest_records (id, full_name, phone, room_number, check_in_date, check_out_date, note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	if guest.ID == uuid.Nil {
		guest.ID = uuid.New()
	}
	now := time.Now().UTC()
	guest.CreatedAt = now
	guest.UpdatedAt = now

	err := r.pool.QueryRow(ctx, query,
		guest.ID,
		guest.FullName,
		guest.Phone,
		guest.RoomNumber,
		guest.CheckInDate,
		guest.CheckOutDate,
		guest.Note,
		guest.CreatedAt,
		guest.UpdatedAt,
	).Scan(&guest.ID, &guest.CreatedAt, &guest.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert guest record: %w", err)
	}

	return nil
}

func (r *guestRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.GuestRecord, error) {
	query := `
		SELECT id, full_name, phone, room_number, check_in_date, check_out_date, note, created_at, updated_at
		FROM guest_records
		WHERE id = $1
	`
	var guest domain.GuestRecord
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&guest.ID,
		&guest.FullName,
		&guest.Phone,
		&guest.RoomNumber,
		&guest.CheckInDate,
		&guest.CheckOutDate,
		&guest.Note,
		&guest.CreatedAt,
		&guest.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get guest record: %w", err)
	}

	return &guest, nil
}

func (r *guestRepo) List(ctx context.Context, filter domain.GuestFilter) ([]domain.GuestRecord, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if strings.TrimSpace(filter.RoomNumber) != "" {
		conditions = append(conditions, fmt.Sprintf("room_number = $%d", argIdx))
		args = append(args, strings.TrimSpace(filter.RoomNumber))
		argIdx++
	}

	if strings.TrimSpace(filter.Phone) != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIdx))
		args = append(args, "%"+strings.TrimSpace(filter.Phone)+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM guest_records %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count guest records: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, full_name, phone, room_number, check_in_date, check_out_date, note, created_at, updated_at
		FROM guest_records
		%s
		ORDER BY check_in_date DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query guest records: %w", err)
	}
	defer rows.Close()

	var records []domain.GuestRecord
	for rows.Next() {
		var g domain.GuestRecord
		err := rows.Scan(
			&g.ID,
			&g.FullName,
			&g.Phone,
			&g.RoomNumber,
			&g.CheckInDate,
			&g.CheckOutDate,
			&g.Note,
			&g.CreatedAt,
			&g.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan guest record: %w", err)
		}
		records = append(records, g)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return records, total, nil
}

func (r *guestRepo) Update(ctx context.Context, guest *domain.GuestRecord) error {
	query := `
		UPDATE guest_records
		SET full_name = $2,
		    phone = $3,
		    room_number = $4,
		    check_in_date = $5,
		    check_out_date = $6,
		    note = $7,
		    updated_at = $8
		WHERE id = $1
	`
	guest.UpdatedAt = time.Now().UTC()

	cmdTag, err := r.pool.Exec(ctx, query,
		guest.ID,
		guest.FullName,
		guest.Phone,
		guest.RoomNumber,
		guest.CheckInDate,
		guest.CheckOutDate,
		guest.Note,
		guest.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update guest record: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *guestRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM guest_records WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete guest record: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
