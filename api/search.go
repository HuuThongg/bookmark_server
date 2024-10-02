package api

import (
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/util"
	"fmt"
	"net/http"

	e "bookmark/api/resource/common/err"

	"github.com/go-chi/chi/v5"
)

func (h *API) SearchLinks(w http.ResponseWriter, r *http.Request) {
	query := chi.URLParam(r, "query")

	q := sqlc.New(h.db)

	payload := r.Context().Value("payload").(*auth.PayLoad)

	percent := "%"

	linkTitle := fmt.Sprintf("%s%s%s", percent, query, percent)

	arg := sqlc.SearchLinkzParams{
		LinkTitle: linkTitle,
		AccountID: payload.AccountID,
	}

	links, err := q.SearchLinkz(r.Context(), arg)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, links)
}

func (h *API) SearchFolders(w http.ResponseWriter, r *http.Request) {

	q := sqlc.New(h.db)

	query := chi.URLParam(r, "query")

	payload := r.Context().Value("payload").(*auth.PayLoad)

	percent := "%"

	folderName := fmt.Sprintf("%s%s%s", percent, query, percent)

	arg := sqlc.SearchFolderzParams{
		//PlaintoTsquery: query,
		FolderName: folderName,
		AccountID:  payload.AccountID,
	}

	folders, err := q.SearchFolderz(r.Context(), arg)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, folders)
}
