-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS bodyfuel;

-- Создаем enum для статуса выполнения упражнения
CREATE TYPE bodyfuel.exercise_status AS ENUM ('pending', 'in_progress', 'completed', 'skipped');

-- Создаем таблицу для связи тренировок и упражнений
CREATE TABLE IF NOT EXISTS bodyfuel.workouts_exercise (
                                                          workout_id UUID NOT NULL,
                                                          exercise_id UUID NOT NULL,
                                                          modify_reps INT NOT NULL,
                                                          modify_relax_time INT NOT NULL,
                                                          calories INT NOT NULL,
                                                          status bodyfuel.exercise_status NOT NULL DEFAULT 'pending',
                                                          updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Создаем индексы для оптимизации запросов
CREATE INDEX idx_workouts_exercise_workout_id
    ON bodyfuel.workouts_exercise(workout_id);

CREATE INDEX idx_workouts_exercise_exercise_id
    ON bodyfuel.workouts_exercise(exercise_id);

CREATE INDEX idx_workouts_exercise_status
    ON bodyfuel.workouts_exercise(status);

CREATE INDEX idx_workouts_exercise_updated_at
    ON bodyfuel.workouts_exercise(updated_at);

-- Составной индекс для частых фильтров
CREATE INDEX idx_workouts_exercise_workout_status
    ON bodyfuel.workouts_exercise(workout_id, status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем индексы
DROP INDEX IF EXISTS bodyfuel.idx_workouts_exercise_workout_status;
DROP INDEX IF EXISTS bodyfuel.idx_workouts_exercise_updated_at;
DROP INDEX IF EXISTS bodyfuel.idx_workouts_exercise_status;
DROP INDEX IF EXISTS bodyfuel.idx_workouts_exercise_exercise_id;
DROP INDEX IF EXISTS bodyfuel.idx_workouts_exercise_workout_id;

-- Удаляем таблицу
DROP TABLE IF EXISTS bodyfuel.workouts_exercise;

-- Удаляем enum
DROP TYPE IF EXISTS bodyfuel.exercise_status;

-- +goose StatementEnd