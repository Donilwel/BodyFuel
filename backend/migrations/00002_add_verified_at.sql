-- +goose Up
ALTER TABLE bodyfuel.user_info
    ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE bodyfuel.user_info
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS phone_verified_at;
