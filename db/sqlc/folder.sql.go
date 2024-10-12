// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: folder.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createFolder = `-- name: CreateFolder :one
INSERT INTO folder (folder_id, folder_name, subfolder_of, account_id, path, label)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

type CreateFolderParams struct {
	FolderID    string      `json:"folder_id"`
	FolderName  string      `json:"folder_name"`
	SubfolderOf pgtype.Text `json:"subfolder_of"`
	AccountID   int64       `json:"account_id"`
	Path        string      `json:"path"`
	Label       string      `json:"label"`
}

func (q *Queries) CreateFolder(ctx context.Context, arg CreateFolderParams) (Folder, error) {
	row := q.db.QueryRow(ctx, createFolder,
		arg.FolderID,
		arg.FolderName,
		arg.SubfolderOf,
		arg.AccountID,
		arg.Path,
		arg.Label,
	)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const deleteFolderForever = `-- name: DeleteFolderForever :one
DELETE FROM folder where path <@ (SELECT path FROM folder where folder.folder_id = $1) RETURNING folder_id
`

func (q *Queries) DeleteFolderForever(ctx context.Context, folderID string) (string, error) {
	row := q.db.QueryRow(ctx, deleteFolderForever, folderID)
	var folder_id string
	err := row.Scan(&folder_id)
	return folder_id, err
}

const getFolder = `-- name: GetFolder :one
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order FROM folder
WHERE folder_id = $1
LIMIT 1
`

func (q *Queries) GetFolder(ctx context.Context, folderID string) (Folder, error) {
	row := q.db.QueryRow(ctx, getFolder, folderID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const getFolderAncestors = `-- name: GetFolderAncestors :many
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order FROM folder
WHERE folder.path @> (
  SELECT path FROM folder as f
  WHERE f.label = $1
)
ORDER BY path
`

func (q *Queries) GetFolderAncestors(ctx context.Context, label string) ([]Folder, error) {
	rows, err := q.db.Query(ctx, getFolderAncestors, label)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFolderByFolderAndAccountIds = `-- name: GetFolderByFolderAndAccountIds :one
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order FROM folder
WHERE folder_id = $1 AND account_id = $2
LIMIT 1
`

type GetFolderByFolderAndAccountIdsParams struct {
	FolderID  string `json:"folder_id"`
	AccountID int64  `json:"account_id"`
}

func (q *Queries) GetFolderByFolderAndAccountIds(ctx context.Context, arg GetFolderByFolderAndAccountIdsParams) (Folder, error) {
	row := q.db.QueryRow(ctx, getFolderByFolderAndAccountIds, arg.FolderID, arg.AccountID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const getFolderNodes = `-- name: GetFolderNodes :many
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order FROM folder
WHERE subfolder_of = $1 AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC
`

func (q *Queries) GetFolderNodes(ctx context.Context, subfolderOf pgtype.Text) ([]Folder, error) {
	rows, err := q.db.Query(ctx, getFolderNodes, subfolderOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFoldersByAccountId = `-- name: GetFoldersByAccountId :many
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
FROM folder
WHERE account_id = $1 AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC
`

func (q *Queries) GetFoldersByAccountId(ctx context.Context, accountID int64) ([]Folder, error) {
	rows, err := q.db.Query(ctx, getFoldersByAccountId, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFoldersMovedToTrash = `-- name: GetFoldersMovedToTrash :many
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order FROM folder WHERE folder_deleted_at IS NOT NULL AND account_id = $1 ORDER BY folder_deleted_at DESC
`

func (q *Queries) GetFoldersMovedToTrash(ctx context.Context, accountID int64) ([]Folder, error) {
	rows, err := q.db.Query(ctx, getFoldersMovedToTrash, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRootFolders = `-- name: GetRootFolders :many
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order FROM folder WHERE NLEVEL(path) = 1 AND account_id = $1 AND folder_deleted_at IS NULL ORDER BY folder_created_at DESC
`

func (q *Queries) GetRootFolders(ctx context.Context, accountID int64) ([]Folder, error) {
	rows, err := q.db.Query(ctx, getRootFolders, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRootNodes = `-- name: GetRootNodes :many
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order FROM folder
WHERE account_id = $1 AND subfolder_of IS NULL AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC
`

func (q *Queries) GetRootNodes(ctx context.Context, accountID int64) ([]Folder, error) {
	rows, err := q.db.Query(ctx, getRootNodes, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTreeFolders = `-- name: GetTreeFolders :many
SELECT folder_id, folder_name, subfolder_of, folder_order
FROM folder
WHERE account_id = $1
ORDER BY folder_order ASC
`

type GetTreeFoldersRow struct {
	FolderID    string      `json:"folder_id"`
	FolderName  string      `json:"folder_name"`
	SubfolderOf pgtype.Text `json:"subfolder_of"`
	FolderOrder int32       `json:"folder_order"`
}

func (q *Queries) GetTreeFolders(ctx context.Context, accountID int64) ([]GetTreeFoldersRow, error) {
	rows, err := q.db.Query(ctx, getTreeFolders, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTreeFoldersRow
	for rows.Next() {
		var i GetTreeFoldersRow
		if err := rows.Scan(
			&i.FolderID,
			&i.FolderName,
			&i.SubfolderOf,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const moveFolder = `-- name: MoveFolder :many
UPDATE folder SET path = (SELECT path FROM folder WHERE folder.label = $1) || SUBPATH(path, NLEVEL((SELECT path FROM folder WHERE folder.label = $2))-1) WHERE path <@ (SELECT path FROM folder WHERE folder.label = $3) RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

type MoveFolderParams struct {
	Label   string `json:"label"`
	Label_2 string `json:"label_2"`
	Label_3 string `json:"label_3"`
}

func (q *Queries) MoveFolder(ctx context.Context, arg MoveFolderParams) ([]Folder, error) {
	rows, err := q.db.Query(ctx, moveFolder, arg.Label, arg.Label_2, arg.Label_3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const moveFolderToTrash = `-- name: MoveFolderToTrash :one
UPDATE folder
SET folder_deleted_at = CURRENT_TIMESTAMP
WHERE folder_id = $1
RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

func (q *Queries) MoveFolderToTrash(ctx context.Context, folderID string) (Folder, error) {
	row := q.db.QueryRow(ctx, moveFolderToTrash, folderID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const moveFoldersToRoot = `-- name: MoveFoldersToRoot :many
UPDATE folder SET path = SUBPATH(path, NLEVEL((SELECT path FROM folder WHERE folder.label = $1))-1) WHERE path <@ (
SELECT path FROM folder WHERE folder.label = $2
) RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

type MoveFoldersToRootParams struct {
	Label   string `json:"label"`
	Label_2 string `json:"label_2"`
}

func (q *Queries) MoveFoldersToRoot(ctx context.Context, arg MoveFoldersToRootParams) ([]Folder, error) {
	rows, err := q.db.Query(ctx, moveFoldersToRoot, arg.Label, arg.Label_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const renameFolder = `-- name: RenameFolder :one
UPDATE folder
SET folder_name = $1
WHERE folder_id = $2
RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

type RenameFolderParams struct {
	FolderName string `json:"folder_name"`
	FolderID   string `json:"folder_id"`
}

func (q *Queries) RenameFolder(ctx context.Context, arg RenameFolderParams) (Folder, error) {
	row := q.db.QueryRow(ctx, renameFolder, arg.FolderName, arg.FolderID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const restoreFolderFromTrash = `-- name: RestoreFolderFromTrash :one
UPDATE folder SET folder_deleted_at = NULL WHERE folder_id = $1 RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

func (q *Queries) RestoreFolderFromTrash(ctx context.Context, folderID string) (Folder, error) {
	row := q.db.QueryRow(ctx, restoreFolderFromTrash, folderID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const searchFolders = `-- name: SearchFolders :many
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
FROM folder
WHERE textsearchable_index_col @@ plainto_tsquery($1) AND account_id = $2 AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC
`

type SearchFoldersParams struct {
	PlaintoTsquery string `json:"plainto_tsquery"`
	AccountID      int64  `json:"account_id"`
}

func (q *Queries) SearchFolders(ctx context.Context, arg SearchFoldersParams) ([]Folder, error) {
	rows, err := q.db.Query(ctx, searchFolders, arg.PlaintoTsquery, arg.AccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchFolderz = `-- name: SearchFolderz :many
SELECT folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
FROM folder
WHERE folder_name ILIKE $1 AND account_id = $2 AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC
`

type SearchFolderzParams struct {
	FolderName string `json:"folder_name"`
	AccountID  int64  `json:"account_id"`
}

func (q *Queries) SearchFolderz(ctx context.Context, arg SearchFolderzParams) ([]Folder, error) {
	rows, err := q.db.Query(ctx, searchFolderz, arg.FolderName, arg.AccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Folder
	for rows.Next() {
		var i Folder
		if err := rows.Scan(
			&i.FolderID,
			&i.AccountID,
			&i.FolderName,
			&i.Path,
			&i.Label,
			&i.Starred,
			&i.FolderCreatedAt,
			&i.FolderUpdatedAt,
			&i.SubfolderOf,
			&i.FolderDeletedAt,
			&i.TextsearchableIndexCol,
			&i.FolderOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const starFolder = `-- name: StarFolder :one
UPDATE folder
SET starred = 'true'
WHERE folder_id = $1
RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

func (q *Queries) StarFolder(ctx context.Context, folderID string) (Folder, error) {
	row := q.db.QueryRow(ctx, starFolder, folderID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const toggleFolderStarred = `-- name: ToggleFolderStarred :one
UPDATE folder SET starred = NOT starred WHERE folder_id = $1 RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

func (q *Queries) ToggleFolderStarred(ctx context.Context, folderID string) (Folder, error) {
	row := q.db.QueryRow(ctx, toggleFolderStarred, folderID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const unstarFolder = `-- name: UnstarFolder :one
UPDATE folder
SET starred = 'false'
WHERE folder_id = $1
RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

func (q *Queries) UnstarFolder(ctx context.Context, folderID string) (Folder, error) {
	row := q.db.QueryRow(ctx, unstarFolder, folderID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const updateFolderOrder = `-- name: UpdateFolderOrder :exec
UPDATE folder
SET folder_order = $1
WHERE folder_id = $2 AND account_id = $3
`

type UpdateFolderOrderParams struct {
	FolderOrder int32  `json:"folder_order"`
	FolderID    string `json:"folder_id"`
	AccountID   int64  `json:"account_id"`
}

func (q *Queries) UpdateFolderOrder(ctx context.Context, arg UpdateFolderOrderParams) error {
	_, err := q.db.Exec(ctx, updateFolderOrder, arg.FolderOrder, arg.FolderID, arg.AccountID)
	return err
}

const updateFolderSubfolderOf = `-- name: UpdateFolderSubfolderOf :one
UPDATE folder
SET subfolder_of = $1
WHERE folder_id = $2
RETURNING folder_id, account_id, folder_name, path, label, starred, folder_created_at, folder_updated_at, subfolder_of, folder_deleted_at, textsearchable_index_col, folder_order
`

type UpdateFolderSubfolderOfParams struct {
	SubfolderOf pgtype.Text `json:"subfolder_of"`
	FolderID    string      `json:"folder_id"`
}

func (q *Queries) UpdateFolderSubfolderOf(ctx context.Context, arg UpdateFolderSubfolderOfParams) (Folder, error) {
	row := q.db.QueryRow(ctx, updateFolderSubfolderOf, arg.SubfolderOf, arg.FolderID)
	var i Folder
	err := row.Scan(
		&i.FolderID,
		&i.AccountID,
		&i.FolderName,
		&i.Path,
		&i.Label,
		&i.Starred,
		&i.FolderCreatedAt,
		&i.FolderUpdatedAt,
		&i.SubfolderOf,
		&i.FolderDeletedAt,
		&i.TextsearchableIndexCol,
		&i.FolderOrder,
	)
	return i, err
}

const updateParentFolderToNull = `-- name: UpdateParentFolderToNull :exec
UPDATE folder
SET subfolder_of = NULL
WHERE folder_id = $1
`

func (q *Queries) UpdateParentFolderToNull(ctx context.Context, folderID string) error {
	_, err := q.db.Exec(ctx, updateParentFolderToNull, folderID)
	return err
}
