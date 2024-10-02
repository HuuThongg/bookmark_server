// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: password_resset_token.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPasswordResetToken = `-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_token (account_id, token_hash, token_expiry) VALUES ($1, $2, $3) ON CONFLICT (account_id) DO UPDATE SET token_hash = EXCLUDED.token_hash, token_expiry = EXCLUDED.token_expiry RETURNING id, account_id, token_hash, token_expiry
`

type CreatePasswordResetTokenParams struct {
	AccountID   int64              `json:"account_id"`
	TokenHash   string             `json:"token_hash"`
	TokenExpiry pgtype.Timestamptz `json:"token_expiry"`
}

func (q *Queries) CreatePasswordResetToken(ctx context.Context, arg CreatePasswordResetTokenParams) (PasswordResetToken, error) {
	row := q.db.QueryRow(ctx, createPasswordResetToken, arg.AccountID, arg.TokenHash, arg.TokenExpiry)
	var i PasswordResetToken
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.TokenHash,
		&i.TokenExpiry,
	)
	return i, err
}

const deletePasswordResetToken = `-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_token WHERE token_hash = $1
`

func (q *Queries) DeletePasswordResetToken(ctx context.Context, tokenHash string) error {
	_, err := q.db.Exec(ctx, deletePasswordResetToken, tokenHash)
	return err
}

const getPasswordResetToken = `-- name: GetPasswordResetToken :one
SELECT id, account_id, token_hash, token_expiry FROM password_reset_token WHERE token_hash = $1 LIMIT 1
`

func (q *Queries) GetPasswordResetToken(ctx context.Context, tokenHash string) (PasswordResetToken, error) {
	row := q.db.QueryRow(ctx, getPasswordResetToken, tokenHash)
	var i PasswordResetToken
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.TokenHash,
		&i.TokenExpiry,
	)
	return i, err
}
