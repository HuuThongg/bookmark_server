-- name: AddTag :one
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
RETURNING *;


-- name: DeleteTag :exec
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
);

-- name: GetTagByLinkId :many
SELECT t.tag_id, t.tag_name
FROM tags t
JOIN link_tags lt ON lt.tag_id = t.tag_id
JOIN link l ON l.link_id = lt.link_id
WHERE lt.link_id = $1 AND l.account_id = $2;

-- name: GetTagsByAccountID :many
SELECT *
FROM tags
WHERE account_id = $1;

-- name: GetTagStatsByAccountID :many
SELECT tag_name, COUNT(*) AS amount
FROM tags
WHERE account_id = $1
GROUP BY tag_name;
