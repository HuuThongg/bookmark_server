package api

import (
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/mailjet"
	"bookmark/util"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	e "bookmark/api/resource/common/err"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type contactSupportRequest struct {
	Message string `json:"message"`
}

func (s contactSupportRequest) validate(reqValidationChan chan error) error {
	returnVal := validation.ValidateStruct(&s,
		validation.Field(&s.Message, validation.Required.When(s.Message == "").Error("message is required")),
	)
	reqValidationChan <- returnVal
	return returnVal
}

func (h *API) ContactSupport(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req contactSupportRequest

	err := body.Decode(&req)
	if err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.validate(reqValidationChan)
	}()

	wg.Wait()

	payload := r.Context().Value("payload").(*auth.PayLoad)

	queries := sqlc.New(h.db)

	account, err := queries.GetAccount(r.Context(), payload.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println(err)
			util.Response(w, "account not found", http.StatusUnauthorized)
			return
		} else {
			log.Println(err)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	newMailRequest := &mailjet.EmailSupportRequest{
		FromEmail: account.Email,
		FromName:  account.Fullname,
		Subject:   req.Message,
		TextPart:  req.Message,
	}
	if err := newMailRequest.EmailSupport(); err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	newMessageParams := sqlc.NewMessageParams{
		Account:     account.ID,
		MessageBody: req.Message,
	}

	if _, err := queries.NewMessage(r.Context(), newMessageParams); err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	util.Response(w, "message sent", http.StatusOK)
}
