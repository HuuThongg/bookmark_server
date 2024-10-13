package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"bookmark/auth"

	"bookmark/db/sqlc"

	"bookmark/util"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgconn"
)

type token struct {
	Token string `json:"token"`
}

func (t token) Validate(requestValidationChan chan error) error {
	err := validation.ValidateStruct(&t,
		validation.Field(&t.Token, validation.Required.Error("token is required")),
	)

	requestValidationChan <- err

	return err
}

func (h *API) CheckIfIsAuthenticated(w http.ResponseWriter, r *http.Request) {

	log := h.logger.With().Str("func", "CheckIfIsAuthenticated").Logger()
	data := json.NewDecoder(r.Body)

	data.DisallowUnknownFields()

	var req token

	if err := data.Decode(&req); err != nil {

		if e, ok := err.(*json.SyntaxError); ok {
			fmt.Println("failed to decode request")
			log.Error().Err(e).Msg("failed to decode request")
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {

			fmt.Println("access_token is empty")
			log.Error().Err(e).Msg("failed to decode request!")
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}
	cacheKey := req.Token // Use the token as the cache key
	ctx := context.Background()
	cachedResponse, err1 := h.redis.Get(ctx, cacheKey).Result()
	if err1 != nil {
		log.Info().Msg("failed to get cache from Redis")
		// util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		// return
	} else {
		log.Info().Msg("CheckIfIsAuthenticated cache hit")
		util.Response(w, cachedResponse, http.StatusOK)
		return
	}
	requestValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(requestValidationChan)
	}()

	wg.Wait()

	err := <-requestValidationChan
	if err != nil {
		if e, ok := err.(validation.InternalError); ok {
			// an internal error happened
			log.Error().Err(e.InternalError()).Msg("an internal sver error occured while validation request body")
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Printf("invalid request: %v", err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	payload, err := auth.VerifyToken(req.Token)
	if err != nil {
		log.Printf("payload good %v", err)
		util.Response(w, err.Error(), http.StatusUnauthorized)
		return
	}

	q := sqlc.New(h.db)
	account, err := q.GetAccount(r.Context(), int64(payload.AccountID))
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Error().Err(err).Msg("failed get account")
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {

			log.Error().Err(err).Msg("account not found,now row")
			util.Response(w, errors.New("account not found").Error(), http.StatusUnauthorized)
			return
		} else {
			log.Error().Err(err).Msg("failed to get account")
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	if account.LastLogin.Time.UTC() != payload.IssuedAt.Time.UTC().Truncate(time.Microsecond) {
		util.Response(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	response := "user logged in"
	if err := h.redis.Set(ctx, cacheKey, response, 10*time.Minute).Err(); err != nil {
		log.Error().Err(err).Msg("failed to set cache in Redis")
	}
	util.Response(w, response, http.StatusOK)
}
