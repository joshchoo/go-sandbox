-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
PRAGMA JOURNAL_MODE=WAL;

CREATE TABLE IF NOT EXISTS blob_cache
(
    id         INTEGER PRIMARY KEY,
    key        TEXT UNIQUE NOT NULL,
    data       BLOB        NOT NULL,
    version    INTEGER     NOT NULL DEFAULT 1,
    created_at INTEGER     NOT NULL DEFAULT (UNIXEPOCH()),
    updated_at INTEGER     NOT NULL DEFAULT (UNIXEPOCH())
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS blob_cache;
-- +goose StatementEnd
