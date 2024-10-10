-- +goose Up
ALTER TABLE link
ADD COLUMN description TEXT DEFAULT '';
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down

ALTER TABLE link
DROP COLUMN IF EXISTS description;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
