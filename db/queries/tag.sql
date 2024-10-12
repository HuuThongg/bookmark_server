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
RETURNING link_id,tag_id;


-- name: DeleteTag :one
WITH deleted_tag AS (
    DELETE FROM link_tags
    WHERE link_tags.link_id = $1 AND link_tags.tag_id = $2
    RETURNING tag_id
),
deleted_tags AS (
    DELETE FROM tags
    WHERE tags.tag_id = (SELECT tag_id FROM deleted_tag) 
    AND NOT EXISTS (
        SELECT 1 
        FROM link_tags 
        WHERE link_tags.tag_id = (SELECT tag_id FROM deleted_tag)
    )
    RETURNING tag_id
)
SELECT tag_id FROM deleted_tag UNION SELECT tag_id FROM deleted_tags;

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
SELECT t.tag_name, COUNT(*) AS amount
FROM tags t
JOIN link_tags lt ON t.tag_id = lt.tag_id
JOIN link l ON lt.link_id = l.link_id
WHERE l.account_id = $1  
GROUP BY t.tag_name;


-- name: AddTags :many
WITH inserted_tags AS (
  INSERT INTO tags (tag_name, account_id)
  SELECT unnest($1::text[]), $2 -- $1 is the array of tag names, $2 is account_id
  -- ON CONFLICT (tag_name, account_id) DO NOTHING
  RETURNING tag_id, tag_name
),
all_tags AS (
  SELECT tag_id, tag_name FROM inserted_tags
  UNION
  SELECT tag_id, tag_name FROM tags WHERE tag_name = ANY($1) AND account_id = $2
)
INSERT INTO link_tags (link_id, tag_id)
SELECT unnest($3::text[]), at.tag_id -- $3 is the array of link_ids
FROM all_tags at
ON CONFLICT (link_id, tag_id) DO NOTHING
RETURNING link_id, tag_id;
