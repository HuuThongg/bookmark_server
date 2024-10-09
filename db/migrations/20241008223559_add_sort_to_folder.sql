-- +goose Up
ALTER TABLE folder
ADD COLUMN folder_order INT NOT NULL DEFAULT 0;
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
ALTER TABLE folder
DROP COLUMN IF EXISTS folder_order;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
