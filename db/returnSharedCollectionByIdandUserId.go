package db

import (
	"context"

	"bookmark/db/connection"

	"bookmark/db/sqlc"
)

func ReturnCollectionMemberByCollectionAndMemberIDs(ctx context.Context, collectionID string, accountID int64) (*sqlc.CollectionMember, error) {
	arg := sqlc.GetCollectionMemberByCollectionAndMemberIDsParams{
		CollectionID: collectionID,
		MemberID:     accountID,
	}
	collectionMember, err := sqlc.New(connection.ConnectDB()).GetCollectionMemberByCollectionAndMemberIDs(ctx, arg)
	if err != nil {
		return &sqlc.CollectionMember{}, err
	}

	return &collectionMember, nil
}
