package db

import (
	"bookmark/db/connection"
	"bookmark/db/sqlc"
	"context"
)

func ReturnFolder(ctx context.Context, folderID string) (*sqlc.Folder, error) {
	q := sqlc.New(connection.ConnectDB())
	folder, err := q.GetFolder(ctx, folderID)
	if err != nil {
		return &sqlc.Folder{}, err
	}
	return &folder, nil
}
