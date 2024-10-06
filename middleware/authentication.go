package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"bookmark/auth"

	"bookmark/db"

	"bookmark/util"
)

/*
Middleware performs some specific function on the HTTP request or response at a specific stage in the HTTP pipeline before or after the user defined controller. Middleware is a design pattern to eloquently add cross cutting concerns like logging, handling authentication without having many code contact points.
*/

func AuthenticateRequest() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			payload, err := getAndVerifyToken(r)
			if err != nil {
				log.Println(err)
				util.Response(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if payload != nil {

				account, err := db.ReturnAccount(r.Context(), payload.AccountID)
				if err != nil {
					log.Println(err)
					util.Response(w, errors.New("unauthorized").Error(), http.StatusUnauthorized)
					return
				}

				if payload.IssuedAt.Time.Unix() != account.LastLogin.Time.Unix() {
					err := errors.New("invalid token")
					log.Println(err)
					util.Response(w, err.Error(), http.StatusUnauthorized)
					return
				}

				ctx := context.WithValue(r.Context(), "payload", payload)

				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}

		return http.HandlerFunc(fn)
	}
}

func getAndVerifyToken(r *http.Request) (*auth.PayLoad, error) {
	token := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	if token == "" {
		return nil, errors.New("token is empty or not in proper format")
	}

	payload, err := auth.VerifyToken(token)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
