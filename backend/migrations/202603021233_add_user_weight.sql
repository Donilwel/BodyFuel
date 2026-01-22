-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS bodyfuel.user_weight (
                                                    id UUID PRIMARY KEY,
                                                    id_user UUID NOT NULL REFERENCES bodyfuel.user_info(id) ON DELETE CASCADE,
                                                    weight FLOAT,
                                                    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_info_username ON bodyfuel.user_info(username);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bodyfuel.user_weight;
-- +goose StatementEnd