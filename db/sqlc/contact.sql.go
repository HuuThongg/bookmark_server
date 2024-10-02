// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: contact.sql

package sqlc

import (
	"context"
)

const getAllMessages = `-- name: GetAllMessages :many
SELECT id, account, message_body FROM contact
`

func (q *Queries) GetAllMessages(ctx context.Context) ([]Contact, error) {
	rows, err := q.db.Query(ctx, getAllMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Contact
	for rows.Next() {
		var i Contact
		if err := rows.Scan(&i.ID, &i.Account, &i.MessageBody); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const newMessage = `-- name: NewMessage :one
INSERT INTO contact (account, message_body)
VALUES ($1, $2)
RETURNING id, account, message_body
`

type NewMessageParams struct {
	Account     int64  `json:"account"`
	MessageBody string `json:"message_body"`
}

func (q *Queries) NewMessage(ctx context.Context, arg NewMessageParams) (Contact, error) {
	row := q.db.QueryRow(ctx, newMessage, arg.Account, arg.MessageBody)
	var i Contact
	err := row.Scan(&i.ID, &i.Account, &i.MessageBody)
	return i, err
}
