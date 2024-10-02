package db

import (
	"bookmark/db/connection"
	"bookmark/db/sqlc"
	"context"
)

func ReturnAccount(ctx context.Context, accountID int64) (*sqlc.Account, error) {
	q := sqlc.New(connection.ConnectDB())
	account, err := q.GetAccount(ctx, accountID)
	if err != nil {
		return &sqlc.Account{}, err
	}
	return &account, nil
}
