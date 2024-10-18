package auth

import (
	"bookmark/db/sqlc"
	"bookmark/util"
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgconn"
)

func LoginUser(account sqlc.Account, q *sqlc.Queries, ctx context.Context) (string, string, http.Cookie) {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("could not load config at login.go: %v", err)
		return "", "", http.Cookie{}
	}

	accessToken, accesstTokenPayLoad, err := CreateToken(account.ID, time.Now().UTC(), config.Access_Token_Duration)
	if err != nil {
		log.Fatalf("could not create access token at login.go: %v", err)
		return "", "", http.Cookie{}
	}

	refreshToken, refreshTokenPayload, err := CreateToken(account.ID, accesstTokenPayLoad.IssuedAt.Time, config.Refresh_Token_Duration)

	if err != nil {
		log.Fatalf("could not refresh access token at login.go: %v", err)
		return "", "", http.Cookie{}
	}
	refreshTokenCookie := http.Cookie{
		Name:     "refreshTokenCookie",
		Value:    refreshToken,
		Path:     "/",
		Expires:  refreshTokenPayload.Expiry.Time,
		Secure:   true,
		SameSite: http.SameSite(http.SameSiteNoneMode),
		HttpOnly: true,
	}
	createAccountSessionParams := sqlc.CreateAccountSessionParams{
		RefreshTokenID: refreshTokenPayload.ID,
		AccountID:      account.ID,
		IssuedAt:       refreshTokenPayload.IssuedAt,
		Expiry:         refreshTokenPayload.Expiry,
		UserAgent:      "",
		ClientIp:       "",
	}
	_, err = q.CreateAccountSession(ctx, createAccountSessionParams)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Fatalf("could not create account session at login.go: %v", pgErr)
			return "", "", http.Cookie{}
		}

		log.Fatalf("could not create account session at auth/login.go: %v", err)
		return "", "", http.Cookie{}
	}

	return accessToken, refreshToken, refreshTokenCookie
}
