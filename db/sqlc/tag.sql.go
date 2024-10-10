// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: tag.sql

package sqlc

import (
	"context"
)

const addTag = `-- name: AddTag :one
WITH inserted_tag AS (
  INSERT INTO tags (tag_name, account_id) 
  VALUES ($2, $3) 
  ON CONFLICT (tag_name) DO NOTHING 
  RETURNING tag_id
), existing_tag AS (
  SELECT tag_id 
  FROM tags 
  WHERE tag_name = $2 AND account_id = $3
)
INSERT INTO link_tags (link_id, tag_id)
SELECT $1, COALESCE(it.tag_id, et.tag_id)
FROM inserted_tag it
FULL JOIN existing_tag et ON true
RETURNING link_id, tag_id
`

type AddTagParams struct {
	LinkID    string `json:"link_id"`
	TagName   string `json:"tag_name"`
	AccountID int64  `json:"account_id"`
}

func (q *Queries) AddTag(ctx context.Context, arg AddTagParams) (LinkTag, error) {
	row := q.db.QueryRow(ctx, addTag, arg.LinkID, arg.TagName, arg.AccountID)
	var i LinkTag
	err := row.Scan(&i.LinkID, &i.TagID)
	return i, err
}

const deleteTag = `-- name: DeleteTag :exec
WITH deleted_tag AS (
    DELETE FROM link_tags
    WHERE link_tags.link_id = $1 AND link_tags.tag_id = $2
    RETURNING tag_id
)
DELETE FROM tags
WHERE tags.tag_id = (
    SELECT tag_id 
    FROM deleted_tag
) 
AND NOT EXISTS (
    SELECT 1 
    FROM link_tags 
    WHERE link_tags.tag_id = (
        SELECT tag_id 
        FROM deleted_tag
    )
)
`

type DeleteTagParams struct {
	LinkID string `json:"link_id"`
	TagID  int32  `json:"tag_id"`
}

func (q *Queries) DeleteTag(ctx context.Context, arg DeleteTagParams) error {
	_, err := q.db.Exec(ctx, deleteTag, arg.LinkID, arg.TagID)
	return err
}

const getTagByLinkId = `-- name: GetTagByLinkId :many
SELECT t.tag_id, t.tag_name
FROM tags t
JOIN link_tags lt ON lt.tag_id = t.tag_id
JOIN link l ON l.link_id = lt.link_id
WHERE lt.link_id = $1 AND l.account_id = $2
`

type GetTagByLinkIdParams struct {
	LinkID    string `json:"link_id"`
	AccountID int64  `json:"account_id"`
}

type GetTagByLinkIdRow struct {
	TagID   int32  `json:"tag_id"`
	TagName string `json:"tag_name"`
}

func (q *Queries) GetTagByLinkId(ctx context.Context, arg GetTagByLinkIdParams) ([]GetTagByLinkIdRow, error) {
	rows, err := q.db.Query(ctx, getTagByLinkId, arg.LinkID, arg.AccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTagByLinkIdRow
	for rows.Next() {
		var i GetTagByLinkIdRow
		if err := rows.Scan(&i.TagID, &i.TagName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTagStatsByAccountID = `-- name: GetTagStatsByAccountID :many
SELECT tag_name, COUNT(*) AS amount
FROM tags
WHERE account_id = $1
GROUP BY tag_name
`

type GetTagStatsByAccountIDRow struct {
	TagName string `json:"tag_name"`
	Amount  int64  `json:"amount"`
}

func (q *Queries) GetTagStatsByAccountID(ctx context.Context, accountID int64) ([]GetTagStatsByAccountIDRow, error) {
	rows, err := q.db.Query(ctx, getTagStatsByAccountID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTagStatsByAccountIDRow
	for rows.Next() {
		var i GetTagStatsByAccountIDRow
		if err := rows.Scan(&i.TagName, &i.Amount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTagsByAccountID = `-- name: GetTagsByAccountID :many
SELECT tag_id, tag_name, account_id
FROM tags
WHERE account_id = $1
`

func (q *Queries) GetTagsByAccountID(ctx context.Context, accountID int64) ([]Tag, error) {
	rows, err := q.db.Query(ctx, getTagsByAccountID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Tag
	for rows.Next() {
		var i Tag
		if err := rows.Scan(&i.TagID, &i.TagName, &i.AccountID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}