package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

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
	// body, err1 := io.ReadAll(r.Body)
	// if err1 != nil {
	// 	log.Printf("failed to read request body: %v", err1)
	// 	util.Response(w, errors.New("failed to read request body").Error(), http.StatusInternalServerError)
	// 	return
	// }
	//
	// // Log the raw request body
	// log.Printf("Request body: %s", string(body))
	//
	// // Reset the body so it can be read again by json.NewDecoder
	// r.Body = io.NopCloser(bytes.NewBuffer(body))

	log := h.logger.With().Str("func", "CheckIfIsAuthenticated").Logger()
	data := json.NewDecoder(r.Body)

	data.DisallowUnknownFields()

	var req token

	if err := data.Decode(&req); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Error().Err(e).Msg("failed to decode request")
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Error().Err(e).Msg("failed to decode request!")
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
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
		log.Println(err)
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

	if account.LastLogin.Time.UTC() != payload.IssuedAt.Time.UTC() {
		util.Response(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	util.Response(w, "user logged in", http.StatusOK)
}
