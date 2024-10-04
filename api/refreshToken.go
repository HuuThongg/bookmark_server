package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"bookmark/auth"

	"bookmark/db/sqlc"
	"bookmark/util"

	"github.com/jackc/pgconn"
)

func (h *API) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With().Str("func", "RefreshToken").Logger()
	c, err := r.Cookie("refreshTokenCookie")
	if err != nil {
		log.Error().Err(err).Msg("refreshTokenCookie is empty")
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	payload, err := auth.VerifyToken(c.Value)
	if err != nil {
		util.Response(w, err.Error(), http.StatusUnauthorized)
		return
	}

	queries := sqlc.New(h.db)

	account, err := queries.GetAccount(r.Context(), payload.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			log.Error().Err(err).Msg("account not found")
			util.Response(w, "account not found", http.StatusUnauthorized)
			return
		}

		log.Error().Err(err).Msg("cannot perform GetAccount")
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	accessToken, accessTokenPayload, err := auth.CreateToken(account.ID, time.Now(), h.config.Access_Token_Duration)
	if err != nil {

		log.Error().Err(err).Msg("failed to create access token")
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, refreshTokenPayload, err := auth.CreateToken(account.ID, accessTokenPayload.IssuedAt.Time, h.config.Refresh_Token_Duration)
	if err != nil {

		log.Error().Err(err).Msg("Faile to create refreshToken")
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	refreshTokenCookie := http.Cookie{
		Name:     "refreshTokenCookie",
		Value:    refreshToken,
		Path:     "/",
		Expires:  refreshTokenPayload.Expiry.Time,
		Secure:   false,
		SameSite: http.SameSiteNoneMode,
		HttpOnly: true,
	}

	http.SetCookie(w, &refreshTokenCookie)

	createAccountSessionParams := sqlc.CreateAccountSessionParams{
		RefreshTokenID: refreshTokenPayload.ID,
		AccountID:      account.ID,
		IssuedAt:       refreshTokenPayload.IssuedAt,
		Expiry:         refreshTokenPayload.Expiry,
		UserAgent:      "",
		ClientIp:       "",
	}

	_, err = queries.CreateAccountSession(r.Context(), createAccountSessionParams)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			// log.Printf("failed co create session with err: %s", pgErr.Message)

			log.Error().Err(err).Msgf("Faile to create refreshToken with err: %s ", pgErr.Message)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Printf("failed to create session with error: %s", err.Error())
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	account, err = queries.GetAccount(r.Context(), account.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("Account not found")
			util.Response(w, "account not found", http.StatusUnauthorized)
			return
		}
		log.Error().Err(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	res := newSession(account, accessToken, refreshToken)

	util.JsonResponse(w, res)
}
