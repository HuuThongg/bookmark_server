-- +goose Up
CREATE TABLE tags (
  tag_id SERIAL PRIMARY KEY,
  tag_name TEXT NOT NULL UNIQUE,
  account_id BIGSERIAL NOT NULL,
  CONSTRAINT fk_account_tags FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE
);

CREATE TABLE link_tags (
  link_id TEXT NOT NULL,
  tag_id INT NOT NULL,
  PRIMARY KEY (link_id, tag_id),
  FOREIGN KEY (link_id) REFERENCES link(link_id) ON DELETE CASCADE,
  FOREIGN KEY (tag_id) REFERENCES tags(tag_id) ON DELETE CASCADE
);
-- +goose StatementBegin
SELECT 'up SQL query executed';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS link_tags;  
DROP TABLE IF EXISTS tags;  

-- +goose StatementBegin
SELECT 'down SQL query executed';
-- +goose StatementEnd
