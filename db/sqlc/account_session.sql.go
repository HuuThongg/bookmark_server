// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: account_session.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createAccountSession = `-- name: CreateAccountSession :one
INSERT INTO account_session (refresh_token_id, account_id, issued_at, expiry, user_agent, client_ip)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (account_id) DO UPDATE SET refresh_token_id = EXCLUDED.refresh_token_id, issued_at = EXCLUDED.issued_at, expiry = EXCLUDED.expiry, user_agent = EXCLUDED.user_agent, client_ip = EXCLUDED.client_ip
RETURNING refresh_token_id, account_id, issued_at, expiry, user_agent, client_ip
`

type CreateAccountSessionParams struct {
	RefreshTokenID string             `json:"refresh_token_id"`
	AccountID      int64              `json:"account_id"`
	IssuedAt       pgtype.Timestamptz `json:"issued_at"`
	Expiry         pgtype.Timestamptz `json:"expiry"`
	UserAgent      string             `json:"user_agent"`
	ClientIp       string             `json:"client_ip"`
}

func (q *Queries) CreateAccountSession(ctx context.Context, arg CreateAccountSessionParams) (AccountSession, error) {
	row := q.db.QueryRow(ctx, createAccountSession,
		arg.RefreshTokenID,
		arg.AccountID,
		arg.IssuedAt,
		arg.Expiry,
		arg.UserAgent,
		arg.ClientIp,
	)
	var i AccountSession
	err := row.Scan(
		&i.RefreshTokenID,
		&i.AccountID,
		&i.IssuedAt,
		&i.Expiry,
		&i.UserAgent,
		&i.ClientIp,
	)
	return i, err
}
