CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS guest_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(32) NOT NULL,
    room_number VARCHAR(32) NOT NULL,
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_dates_valid CHECK (check_out_date > check_in_date)
);

CREATE INDEX IF NOT EXISTS idx_guest_records_room_dates 
    ON guest_records (room_number, check_in_date, check_out_date);

CREATE INDEX IF NOT EXISTS idx_guest_records_phone 
    ON guest_records (phone);

CREATE INDEX IF NOT EXISTS idx_guest_records_check_in_date 
    ON guest_records (check_in_date);
