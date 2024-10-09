package api

import (
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/util"
	"net/http"

	e "bookmark/api/resource/common/err"
)

func (h *API) GetTagStats(w http.ResponseWriter, r *http.Request) {

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	tags, err := q.GetTagStatsByAccountID(r.Context(), payload.AccountID)
	if err != nil {
		h.logger.Error().Err(err).Msg("cannot get tags by link id")
		e.ErrorInternalServer(w, err)
		return
	}
	response := map[string]interface{}{
		"tags":   tags,
		"result": true,
	}

	util.JsonResponse(w, response)
}
