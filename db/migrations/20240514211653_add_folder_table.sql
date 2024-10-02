-- +goose Up
CREATE EXTENSION IF NOT EXISTS ltree;
DROP TABLE IF EXISTS folder;

CREATE TABLE folder (
  folder_id TEXT NOT NULL PRIMARY KEY,
  account_id BIGSERIAL NOT NULL,
  folder_name TEXT NOT NULL,
  path ltree NOT NULl,
    label TEXT CHECK (label ~* '^[A-Za-z0-9_]{1,255}$') NOT NULL UNIQUE,
  starred BOOLEAN NOT NULL DEFAULT 'false',
  folder_created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  folder_updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  subfolder_of TEXT,
  folder_deleted_at TIMESTAMPTZ,
  CONSTRAINT fk_folder FOREIGN KEY (subfolder_of) REFERENCES folder (folder_id) ON DELETE CASCADE,
  CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
  UNIQUE(folder_id, folder_name)
);

ALTER TABLE folder ADD COLUMN textsearchable_index_col tsvector GENERATED ALWAYS AS (to_tsvector('english', folder_name)) STORED;

CREATE INDEX IF NOT EXISTS path_gist_idx ON folder USING GIST (path);

CREATE INDEX IF NOT EXISTS folder_name_search_idx ON folder USING GIN (textsearchable_index_col);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
ALTER TABLE IF EXISTS folder DROP CONSTRAINT IF EXISTS fk_folder;
ALTER TABLE IF EXISTS folder DROP CONSTRAINT IF EXISTS fk_account;
DROP TABLE IF EXISTS folder;
