-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS bodyfuel;

DO
$$
BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_type
                       WHERE typname = 'workouts_level'
                         AND typnamespace = 'bodyfuel'::regnamespace) THEN
CREATE TYPE bodyfuel.workouts_level AS ENUM (
                'workout_light',
                'workout_middle',
                'workout_hard'
                );
END IF;
END
$$;

DO
$$
BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_type
                       WHERE typname = 'workouts_status'
                         AND typnamespace = 'bodyfuel'::regnamespace) THEN
CREATE TYPE bodyfuel.workouts_status AS ENUM (
                'workout_created',
                'workout_done',
                'workout_in_active',
                'workout_failed'
                );
END IF;
END
$$;

CREATE TABLE IF NOT EXISTS bodyfuel.workout (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    level bodyfuel.workouts_level NOT NULL,
    status bodyfuel.workouts_status NOT NULL,
    total_calories INT NOT NULL,
    prediction_calories INT NOT NULL,
    duration INTERVAL NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
    );

CREATE INDEX IF NOT EXISTS idx_workout_user_id ON bodyfuel.workout(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_status ON bodyfuel.workout(status);
CREATE INDEX IF NOT EXISTS idx_workout_level ON bodyfuel.workout(level);
CREATE INDEX IF NOT EXISTS idx_workout_created_at ON bodyfuel.workout(created_at);
CREATE INDEX IF NOT EXISTS idx_workout_user_status ON bodyfuel.workout(user_id, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS bodyfuel.idx_workout_user_id;
DROP INDEX IF EXISTS bodyfuel.idx_workout_status;
DROP INDEX IF EXISTS bodyfuel.idx_workout_level;
DROP INDEX IF EXISTS bodyfuel.idx_workout_created_at;
DROP INDEX IF EXISTS bodyfuel.idx_workout_user_status;
DROP INDEX IF EXISTS bodyfuel.idx_workout_exercises_gin;

DROP TABLE IF EXISTS bodyfuel.workout;

DROP TYPE IF EXISTS bodyfuel.exercise_status;
DROP TYPE IF EXISTS bodyfuel.workouts_status;
DROP TYPE IF EXISTS bodyfuel.workouts_level;

-- +goose StatementEnd