-- +goose Up
CREATE SCHEMA IF NOT EXISTS bodyfuel;

CREATE TABLE IF NOT EXISTS bodyfuel.user_info (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT,
    surname TEXT,
    password TEXT NOT NULL,
    email TEXT UNIQUE,
    phone TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_info ON bodyfuel.user_info(username);


-- +goose Down
DROP SCHEMA IF EXISTS bodyfuel CASCADE;



