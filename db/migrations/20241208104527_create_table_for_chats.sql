-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "chats" (
    chat_id BIGINT NOT NULL PRIMAry KEY,
	 type VARCHAR(255) NOT NULL,
	 title VARCHAR(255) NOT NULL
);

ALTER TABLE "messages" ADD COLUMN chat_id BIGINT;

ALTER TABLE "messages" ADD CONSTRAINT fk_chat FOREIGN KEY (chat_id) REFERENCES "chats" (chat_id) ON DELETE CASCADE;

ALTER TABLE "messages" DROP CONSTRAINT IF EXISTS messages_pkey;

ALTER TABLE "messages" ADD CONSTRAINT messages_pkey PRIMARY KEY (message_id, chat_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chats;
-- +goose StatementEnd
