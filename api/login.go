package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"bookmark/auth"

	"bookmark/db/sqlc"
	"bookmark/util"

	e "bookmark/api/resource/common/err"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/jackc/pgconn"
)

type signup struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func (s signup) Validate(requestValidationChan chan error) error {
	err := validation.ValidateStruct(&s,
		validation.Field(&s.Email, validation.Required.Error("email address is required"), is.Email.Error("email must be valid email address")),
		validation.Field(&s.Password, validation.Required.Error("password is required"), validation.Length(6, 1000).Error("password must be at least 6 characters long")),
		validation.Field(&s.Username, validation.Required.Error("Username is required")),
	)

	requestValidationChan <- err

	return err
}

func (h *API) SignUp(w http.ResponseWriter, r *http.Request) {
	data := json.NewDecoder(r.Body)

	data.DisallowUnknownFields()

	var req signup

	if err := data.Decode(&req); err != nil {
		e.ErrorDecodingRequest(w, err)
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
		e.ErrorInternalServer(w, err)
		return
	}

	q := sqlc.New(h.db)
	email := req.Email

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Println(err)
		return
	}

	existingAccount, err := q.GetAccountByEmail(r.Context(), email)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			e.ErrorInternalServer(w, pgErr)
			return
		case errors.Is(err, sql.ErrNoRows):
			createAccountParams := sqlc.NewAccountParams{
				Fullname:        req.Username,
				Email:           req.Email,
				AccountPassword: req.Password,
				// Picture:  req.Picture,
			}
			createAccount(createAccountParams, q, w, h, config, r.Context())
			return
		default:
			e.ErrorInternalServer(w, err)
			return
		}
	}
	if existingAccount != (sqlc.Account{}) {
		loginUser(existingAccount, w, h, config, r.Context())
	}

}

type signin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s signin) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Email, validation.Required.Error("email address is required"), is.Email.Error("email must be a valid email address")),
		validation.Field(&s.Password, validation.Required.Error("password is required"), validation.Length(6, 1000).Error("password must be at least 6 characters long")),
	)
}

func (h *API) SignIn(w http.ResponseWriter, r *http.Request) {
	defer util.TrackTime(time.Now(), "SignIn")

	start := time.Now()
	log.Printf("request body: %v", r.Body)
	data := json.NewDecoder(r.Body)

	data.DisallowUnknownFields()

	var req signin

	if err := data.Decode(&req); err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	// Measure validation time
	if err := req.Validate(); err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	// Measure time for querying account
	q := sqlc.New(h.db)
	account, err := q.GetAccountByEmail(r.Context(), req.Email)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	if !util.CompareHash(req.Password, account.AccountPassword) {
		log.Println("invalid password")
		util.Response(w, "invalid password", http.StatusUnauthorized)
		return
	}

	// Load configuration
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Printf("failed to load config file with err: %v", err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	// Measure token generation time
	accessToken, accessTokenPayload, err := auth.CreateToken(account.ID, time.Now().UTC(), config.Access_Token_Duration)
	if err != nil {
		log.Printf("failed to create access token with err: %v", err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}
	refreshToken, refreshTokenPayload, err := auth.CreateToken(account.ID, accessTokenPayload.IssuedAt.Time, config.Refresh_Token_Duration)
	if err != nil {
		log.Printf("failed to create refresh token with err: %v", err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	// Set refresh token cookie
	refreshTokenCookie := http.Cookie{
		Name:     "refreshTokenCookie",
		Value:    refreshToken,
		Path:     "/",
		Expires:  refreshTokenPayload.Expiry.Time,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	}
	http.SetCookie(w, &refreshTokenCookie)

	// Measure session creation time
	createAccountSessionParams := sqlc.CreateAccountSessionParams{
		RefreshTokenID: refreshTokenPayload.ID,
		AccountID:      account.ID,
		IssuedAt:       refreshTokenPayload.IssuedAt,
		Expiry:         refreshTokenPayload.Expiry,
		UserAgent:      "",
		ClientIp:       "",
	}
	_, err = q.CreateAccountSession(r.Context(), createAccountSessionParams)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	account, err = q.GetAccount(r.Context(), account.ID)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	res := newSession(account, accessToken, refreshToken)
	http.SetCookie(w, &http.Cookie{
		Name:     "is_authenticate",
		Value:    "true",
		Path:     "/",
		HttpOnly: false, // Prevent JavaScript access
		Secure:   false, // Use only on HTTPS
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(24 * time.Hour), // Set expiration
	})
	util.JsonResponse(w, res)
	log.Printf("JsonResponse took %s", time.Since(start))
}
