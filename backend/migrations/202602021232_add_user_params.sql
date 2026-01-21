-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS bodyfuel;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_type
                       WHERE typname = 'want'
                         AND typnamespace = 'bodyfuel'::regnamespace) THEN
            CREATE TYPE bodyfuel.want AS ENUM (
                'lose_weight',
                'build_muscle',
                'stay_fit'
                );
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_type
                       WHERE typname = 'lifestyle'
                         AND typnamespace = 'bodyfuel'::regnamespace) THEN
            CREATE TYPE bodyfuel.lifestyle AS ENUM (
                'not_active',
                'active',
                'sportive'
                );
        END IF;
    END
$$;

-- 3. Создаём таблицу user_params
CREATE TABLE IF NOT EXISTS bodyfuel.user_params (
                                                    id UUID PRIMARY KEY,
                                                    id_user UUID NOT NULL REFERENCES bodyfuel.user_info(id) ON DELETE CASCADE,
                                                    height INT,
                                                    photo TEXT,
                                                    wants bodyfuel.want,
                                                    lifestyle bodyfuel.lifestyle
);

-- 4. Создаём индекс, если его нет
CREATE INDEX IF NOT EXISTS idx_user_info_username ON bodyfuel.user_info(username);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS bodyfuel.user_params;

DROP TYPE IF EXISTS bodyfuel.want;
DROP TYPE IF EXISTS bodyfuel.lifestyle;

DROP INDEX IF EXISTS bodyfuel.idx_user_info_username;

-- +goose StatementEnd
