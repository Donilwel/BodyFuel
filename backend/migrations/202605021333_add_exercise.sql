-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS bodyfuel;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_type
                       WHERE typname = 'level_preparation'
                         AND typnamespace = 'bodyfuel'::regnamespace) THEN
            CREATE TYPE bodyfuel.level_preparation AS ENUM (
                'beginner',
                'medium',
                'sportsman'
                );
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_type
                       WHERE typname = 'exercise_type'
                         AND typnamespace = 'bodyfuel'::regnamespace) THEN
            CREATE TYPE bodyfuel.exercise_type AS ENUM (
                'cardio',
                'upper_body',
                'lower_body',
                'full_body',
                'flexibility'
                );
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_type
                       WHERE typname = 'place_exercise'
                         AND typnamespace = 'bodyfuel'::regnamespace) THEN
            CREATE TYPE bodyfuel.place_exercise AS ENUM (
                'home',
                'gym',
                'street'
                );
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS bodyfuel.exercise (
                                                 id UUID PRIMARY KEY,
                                                 level_preparation bodyfuel.level_preparation NOT NULL,
                                                 name VARCHAR(100) NOT NULL,
                                                 type_exercise bodyfuel.exercise_type NOT NULL,
                                                 description TEXT,
                                                 base_count_reps INT NOT NULL,
                                                 steps INT NOT NULL,
                                                 link_gif TEXT,
                                                 place_exercise bodyfuel.place_exercise NOT NULL,
                                                 avg_calories_per DECIMAL(5,2) NOT NULL,
                                                 base_relax_time INT NOT NULL
                                             );

CREATE INDEX IF NOT EXISTS idx_exercise_level ON bodyfuel.exercise(level_preparation);
CREATE INDEX IF NOT EXISTS idx_exercise_type ON bodyfuel.exercise(type_exercise);
CREATE INDEX IF NOT EXISTS idx_exercise_place ON bodyfuel.exercise(place_exercise);
CREATE INDEX IF NOT EXISTS idx_exercise_name ON bodyfuel.exercise(name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS bodyfuel.idx_exercise_level;
DROP INDEX IF EXISTS bodyfuel.idx_exercise_type;
DROP INDEX IF EXISTS bodyfuel.idx_exercise_place;
DROP INDEX IF EXISTS bodyfuel.idx_exercise_name;

DROP TABLE IF EXISTS bodyfuel.exercise;

DROP TYPE IF EXISTS bodyfuel.level_preparation;
DROP TYPE IF EXISTS bodyfuel.exercise_type;
DROP TYPE IF EXISTS bodyfuel.place_exercise;

-- +goose StatementEnd