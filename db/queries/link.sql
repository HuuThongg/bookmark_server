-- name: AddLink :one
INSERT INTO link (link_id, link_title, link_hostname, link_url, link_favicon, account_id, folder_id, link_thumbnail ,description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetRootLinks :many
SELECT * FROM link WHERE account_id = $1 AND folder_id IS NULL AND deleted_at IS NULL ORDER BY added_at DESC;

-- name: GetAllLinks :many
SELECT 
    l.link_id, 
    l.link_title, 
    l.link_thumbnail, 
    l.link_favicon, 
    l.link_hostname, 
    l.link_url, 
    l.link_notes, 
    l.account_id, 
    l.folder_id, 
    l.added_at, 
    l.updated_at, 
    l.deleted_at, 
    l.description,
    f.folder_name,
JSON_AGG(
        CASE 
            WHEN t.tag_name IS NOT NULL AND t.tag_id IS NOT NULL 
            THEN JSON_BUILD_OBJECT('tag_name', t.tag_name, 'tag_id', t.tag_id)
            ELSE NULL
        END
    ) FILTER (WHERE t.tag_name IS NOT NULL AND t.tag_id IS NOT NULL) AS tags
FROM 
    link l
LEFT JOIN 
    folder f ON l.folder_id = f.folder_id
LEFT JOIN 
    link_tags lt ON l.link_id = lt.link_id
LEFT JOIN 
    tags t ON lt.tag_id = t.tag_id
WHERE 
    l.account_id = $1 AND l.deleted_at IS NULL
GROUP BY 
    l.link_id, l.link_title, l.link_thumbnail, l.link_favicon, 
    l.link_hostname, l.link_url, l.link_notes, l.account_id, 
    l.folder_id, l.added_at, l.updated_at, l.deleted_at, 
    l.description, f.folder_name
ORDER BY 
    l.added_at DESC;


-- name: GetFolderLinks :many
SELECT 
    l.link_id, 
    l.link_title, 
    l.link_thumbnail, 
    l.link_favicon, 
    l.link_hostname, 
    l.link_url, 
    l.link_notes, 
    l.account_id, 
    l.folder_id, 
    l.added_at, 
    l.updated_at, 
    l.deleted_at, 
    l.description,
    f.folder_name,
JSON_AGG(
        CASE 
            WHEN t.tag_name IS NOT NULL AND t.tag_id IS NOT NULL 
            THEN JSON_BUILD_OBJECT('tag_name', t.tag_name, 'tag_id', t.tag_id)
            ELSE NULL
        END
    ) FILTER (WHERE t.tag_name IS NOT NULL AND t.tag_id IS NOT NULL) AS tags
FROM 
    link l
LEFT JOIN 
    folder f ON l.folder_id = f.folder_id
LEFT JOIN 
    link_tags lt ON l.link_id = lt.link_id
LEFT JOIN 
    tags t ON lt.tag_id = t.tag_id
WHERE 
    l.folder_id = $1 AND l.deleted_at IS NULL
GROUP BY 
    l.link_id, l.link_title, l.link_thumbnail, l.link_favicon, 
    l.link_hostname, l.link_url, l.link_notes, l.account_id, 
    l.folder_id, l.added_at, l.updated_at, l.deleted_at, 
    l.description, f.folder_name
ORDER BY 
    l.added_at DESC;

-- name: RenameLink :one
UPDATE link SET link_title = $1 WHERE link_id = $2 RETURNING *;

-- name: MoveLinkToFolder :one
UPDATE link SET folder_id = $1 WHERE link_id = $2 RETURNING *;

-- name: MoveLinkToRoot :one
UPDATE link SET folder_id = NULL WHERE link_id = $1 RETURNING *;

-- name: MoveLinkToTrash :one
UPDATE link SET deleted_at = CURRENT_TIMESTAMP WHERE link_id = $1 RETURNING *;

-- name: RestoreLinkFromTrash :one
UPDATE link SET deleted_at = NULL WHERE link_id = $1 RETURNING *;

-- name: GetLinksMovedToTrash :many
SELECT * FROM link WHERE deleted_at IS NOT NULL AND account_id = $1 ORDER BY deleted_at DESC;

-- name: DeleteLinkForever :one
DELETE FROM link WHERE link_id = $1 RETURNING *;

-- name: SearchLinks :many
SELECT *
FROM link
WHERE textsearchable_index_col @@ plainto_tsquery($1) AND account_id = $2 AND deleted_at IS NULL
ORDER BY added_at DESC;

-- name: SearchLinkz :many
SELECT *
FROM link
WHERE link_title ILIKE $1 AND account_id = $2 AND deleted_at IS NULL
ORDER BY added_at DESC;

-- name: GetLink :one
SELECT * FROM link
WHERE link_id = $1 AND account_id = $2
LIMIT 1;

-- name: GetLinksByUserID :many
SELECT * FROM link WHERE account_id = $1;


-- name: AddNote :one
UPDATE link
SET link_notes = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE link_id = $1 AND account_id = $3
RETURNING link_id, link_notes;

-- name: ChangeLinkTitle :one
UPDATE link
SET link_title = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE link_id = $1 AND account_id = $3
RETURNING link_title;

-- name: ChangeLinkURL :one
UPDATE link
SET link_url = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE link_id = $1 AND account_id = $3
RETURNING link_url;

-- name: GetAllDeletedLinks :many
SELECT * FROM link 
WHERE account_id = $1 AND deleted_at IS NOT NULL;

-- name: GetLinksByTagName :many
SELECT l.*
FROM link l
JOIN link_tags lt ON l.link_id = lt.link_id
JOIN tags t ON lt.tag_id = t.tag_id
WHERE t.tag_name = $1 AND l.deleted_at IS NULL AND l.account_id = $2;

-- name: UpdateLinkDesc :one
UPDATE link
SET description = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE link_id = $1 AND account_id = $3
RETURNING description;
