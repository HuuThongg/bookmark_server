package db

import (
	"context"

	"bookmark/db/connection"

	"bookmark/db/sqlc"
)

func ReturnAncestorsOfFolder(ctx context.Context, folderID string) ([]sqlc.Folder, error) {
	q := sqlc.New(connection.ConnectDB())

	folder, err := q.GetFolder(ctx, folderID)
	if err != nil {
		return []sqlc.Folder{}, err
	}

	ancestorsOfFolder, err := q.GetFolderAncestors(ctx, folder.Label)
	if err != nil {
		return []sqlc.Folder{}, err
	}

	return ancestorsOfFolder, err
}
