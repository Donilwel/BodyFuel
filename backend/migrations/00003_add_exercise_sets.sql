-- +goose Up
ALTER TABLE bodyfuel.workouts_exercise
    ADD COLUMN IF NOT EXISTS sets INT NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE bodyfuel.workouts_exercise
    DROP COLUMN IF EXISTS sets;
