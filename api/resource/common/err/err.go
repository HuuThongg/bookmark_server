package err

import (
	"bookmark/util"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgconn"
)

var (
	RespDBDataInsertFailure = []byte(`{"error": "db data insert failure"}`)
	RespDBDataAccessFailure = []byte(`{"error": "db data access failure"}`)
	RespDBDataUpdateFailure = []byte(`{"error": "db data update failure"}`)
	RespDBDataRemoveFailure = []byte(`{"error": "db data remove failure"}`)

	RespJSONEncodeFailure = []byte(`{"error": "json encode failure"}`)
	RespJSONDecodeFailure = []byte(`{"error": "json decode failure"}`)

	RespInvalidURLParamID = []byte(`{"error": "invalid url param-id"}`)
)

type Error struct {
	Error string `json:"error"`
}

type Errors struct {
	Errors []string `json:"errors"`
}

func ServerError(w http.ResponseWriter, error []byte) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(error)
}

func BadRequest(w http.ResponseWriter, error []byte) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write(error)
}

func ValidationErrors(w http.ResponseWriter, reps []byte) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write(reps)
}

var (
	internalServerError = "something went wrong"
	badRequest          = "bad request"
)

func ErrorPgError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		log.Println(pgErr)
		util.Response(w, internalServerError, http.StatusInternalServerError)
	} else {
		log.Print(err)
		util.Response(w, internalServerError, http.StatusInternalServerError)
	}
}

func ErrorInternalServer(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr):
		log.Println(pgErr)
		util.Response(w, internalServerError, http.StatusInternalServerError)
	case errors.Is(err, sql.ErrNoRows):
		log.Println(sql.ErrNoRows.Error())
		util.Response(w, "not found", http.StatusNotFound)
	default:
		log.Println(err)
		util.Response(w, internalServerError, http.StatusInternalServerError)
	}
}

func ErrorDecodingRequest(w http.ResponseWriter, err error) {
	if e, ok := err.(*json.SyntaxError); ok {
		log.Printf("JSON syntax error occurred at offset byte: %d", e.Offset)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
	} else {
		log.Printf("error decoding request body to struct: %v", err)
		util.Response(w, "bad request", http.StatusBadRequest)
	}
}

func ErrorInvalidRequest(w http.ResponseWriter, err error) {
	if e, ok := err.(validation.InternalError); ok {
		log.Println(e)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
	} else {
		log.Println(err)
		util.Response(w, err.Error(), http.StatusInternalServerError)
	}
}
