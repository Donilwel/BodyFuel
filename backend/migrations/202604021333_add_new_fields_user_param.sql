-- +goose Up
ALTER TABLE bodyfuel.user_params
    ADD COLUMN IF NOT EXISTS target_workouts_weeks INT,
    ADD COLUMN IF NOT EXISTS target_calories_daily INT,
    ADD COLUMN IF NOT EXISTS target_weight FLOAT;

-- +goose Down
ALTER TABLE bodyfuel.user_params
    DROP COLUMN IF EXISTS target_workouts_weeks,
    DROP COLUMN IF EXISTS target_calories_daily,
    DROP COLUMN IF EXISTS target_weight;