// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: account.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const emailExists = `-- name: EmailExists :one
SELECT EXISTS ( SELECT id, fullname, email, email_verified, picture, account_password, create_at, intention, last_login FROM account WHERE email = $1 LIMIT 1)
`

func (q *Queries) EmailExists(ctx context.Context, email string) (bool, error) {
	row := q.db.QueryRow(ctx, emailExists, email)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const getAccount = `-- name: GetAccount :one
SELECT id, fullname, email, email_verified, picture, account_password, create_at, intention, last_login FROM account 
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetAccount(ctx context.Context, id int64) (Account, error) {
	row := q.db.QueryRow(ctx, getAccount, id)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Fullname,
		&i.Email,
		&i.EmailVerified,
		&i.Picture,
		&i.AccountPassword,
		&i.CreateAt,
		&i.Intention,
		&i.LastLogin,
	)
	return i, err
}

const getAccountByEmail = `-- name: GetAccountByEmail :one
SELECT id, fullname, email, email_verified, picture, account_password, create_at, intention, last_login FROM account 
WHERE email = $1
LIMIT 1
`

func (q *Queries) GetAccountByEmail(ctx context.Context, email string) (Account, error) {
	row := q.db.QueryRow(ctx, getAccountByEmail, email)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Fullname,
		&i.Email,
		&i.EmailVerified,
		&i.Picture,
		&i.AccountPassword,
		&i.CreateAt,
		&i.Intention,
		&i.LastLogin,
	)
	return i, err
}

const getAccountLastLogin = `-- name: GetAccountLastLogin :one
SELECT DATE(last_login) FROM account WHERE id = $1 LIMIT 1
`

func (q *Queries) GetAccountLastLogin(ctx context.Context, id int64) (pgtype.Date, error) {
	row := q.db.QueryRow(ctx, getAccountLastLogin, id)
	var date pgtype.Date
	err := row.Scan(&date)
	return date, err
}

const getAllAccounts = `-- name: GetAllAccounts :many
SELECT id, fullname, email, email_verified, picture, account_password, create_at, intention, last_login FROM account
`

func (q *Queries) GetAllAccounts(ctx context.Context) ([]Account, error) {
	rows, err := q.db.Query(ctx, getAllAccounts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.Fullname,
			&i.Email,
			&i.EmailVerified,
			&i.Picture,
			&i.AccountPassword,
			&i.CreateAt,
			&i.Intention,
			&i.LastLogin,
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

const newAccount = `-- name: NewAccount :one
INSERT INTO account (fullname, email, account_password,picture)
VALUES ($1, $2, $3, $4)
RETURNING id, fullname, email, email_verified, picture, account_password, create_at, intention, last_login
`

type NewAccountParams struct {
	Fullname        string `json:"fullname"`
	Email           string `json:"email"`
	AccountPassword string `json:"account_password"`
	Picture         string `json:"picture"`
}

func (q *Queries) NewAccount(ctx context.Context, arg NewAccountParams) (Account, error) {
	row := q.db.QueryRow(ctx, newAccount,
		arg.Fullname,
		arg.Email,
		arg.AccountPassword,
		arg.Picture,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Fullname,
		&i.Email,
		&i.EmailVerified,
		&i.Picture,
		&i.AccountPassword,
		&i.CreateAt,
		&i.Intention,
		&i.LastLogin,
	)
	return i, err
}

const updateAccountEmailVerificationStatus = `-- name: UpdateAccountEmailVerificationStatus :exec
Update account SET email_verified = 'TRUE' WHERE email = $1
`

func (q *Queries) UpdateAccountEmailVerificationStatus(ctx context.Context, email string) error {
	_, err := q.db.Exec(ctx, updateAccountEmailVerificationStatus, email)
	return err
}

const updateLastLogin = `-- name: UpdateLastLogin :one
UPDATE account
SET last_login = $1
WHERE id = $2
RETURNING id, fullname, email, email_verified, picture, account_password, create_at, intention, last_login
`

type UpdateLastLoginParams struct {
	LastLogin pgtype.Timestamptz `json:"last_login"`
	ID        int64              `json:"id"`
}

func (q *Queries) UpdateLastLogin(ctx context.Context, arg UpdateLastLoginParams) (Account, error) {
	row := q.db.QueryRow(ctx, updateLastLogin, arg.LastLogin, arg.ID)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Fullname,
		&i.Email,
		&i.EmailVerified,
		&i.Picture,
		&i.AccountPassword,
		&i.CreateAt,
		&i.Intention,
		&i.LastLogin,
	)
	return i, err
}

const updatePassword = `-- name: UpdatePassword :exec
UPDATE account SET account_password = $1 WHERE id = $2
`

type UpdatePasswordParams struct {
	AccountPassword string `json:"account_password"`
	ID              int64  `json:"id"`
}

func (q *Queries) UpdatePassword(ctx context.Context, arg UpdatePasswordParams) error {
	_, err := q.db.Exec(ctx, updatePassword, arg.AccountPassword, arg.ID)
	return err
}
