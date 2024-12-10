-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "messages" (
    message_id INTEGER PRIMARY KEY,
    user_full_name VARCHAR NOT NULL,
    text VARCHAR NOT NULL,
    date TIMESTAMP NOT NULL,
    label INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "messages";
-- +goose StatementEnd
