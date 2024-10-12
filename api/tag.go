package api

import (
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/util"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	e "bookmark/api/resource/common/err"

	"github.com/go-playground/validator/v10"
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

type DeleteTagReq struct {
	LinkID string `json:"link_id" validate:"required,max=1000"`
	TagID  string `json:"tag_id" validate:"required"`
}

func (h *API) DeleteTag(w http.ResponseWriter, r *http.Request) {

	var deleteTagReq DeleteTagReq
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()
	if err := body.Decode(&deleteTagReq); err != nil {
		h.logger.Error().Err(err).Msg("Cannot decode ChangeLinkURL")
		e.ErrorDecodingRequest(w, err)
		return
	}
	if err := h.validator.Struct(deleteTagReq); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Validation error")
		}
		util.JsonResponse(w, "Validation failed", http.StatusBadRequest)
		return
	}
	tagId, err := strconv.Atoi(deleteTagReq.TagID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// payload := r.Context().Value("payload").(*auth.PayLoad)
	q := sqlc.New(h.db)
	params := sqlc.DeleteTagParams{
		LinkID: deleteTagReq.LinkID,
		TagID:  int32(tagId),
	}
	fmt.Println("linkid", deleteTagReq.LinkID)
	fmt.Println("tagId", int32(tagId))
	tag_id, err := q.DeleteTag(r.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("cannot delete tags")
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, tag_id)
}

type AddTagReq struct {
	LinkID  string `json:"link_id" validate:"required,max=1000"`
	TagName string `json:"tag_name" validate:"required"`
}

func (h *API) AddTag(w http.ResponseWriter, r *http.Request) {

	var addTagReq AddTagReq
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()
	if err := body.Decode(&addTagReq); err != nil {
		h.logger.Error().Err(err).Msg("Cannot decode ChangeLinkURL")
		e.ErrorDecodingRequest(w, err)
		return
	}
	if err := h.validator.Struct(addTagReq); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Validation error")
		}
		util.JsonResponse(w, "Validation failed", http.StatusBadRequest)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)
	q := sqlc.New(h.db)
	params := sqlc.AddTagParams{
		LinkID:    addTagReq.LinkID,
		TagName:   addTagReq.TagName,
		AccountID: payload.AccountID,
	}

	tag, err := q.AddTag(r.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("cannot add tag")
		e.ErrorInternalServer(w, err)
		return
	}

	response := map[string]interface{}{
		"tag":    tag,
		"result": true,
	}

	util.JsonResponse(w, response)
}

type Tag struct {
	TagID   int64  `json:"tag_id"`
	TagName string `json:"tag_name"`
}

func (h *API) GetTagByLinkId(w http.ResponseWriter, r *http.Request) {
	linkID := r.URL.Query().Get("link_id")
	if linkID == "" {
		util.JsonResponse(w, "link_id is required", http.StatusBadRequest)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)
	params := sqlc.GetTagByLinkIdParams{
		LinkID:    linkID,
		AccountID: payload.AccountID,
	}

	tags, err := q.GetTagByLinkId(r.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("cannot get tags by link id")
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, tags)
}

type AddTagsForLinksReq struct {
	LinkIDs  []string `json:"linkIds" validate:"required,max=1000"`
	TagNames []string `json:"tags" validate:"required"`
}

func (h *API) AddTagsForLinks(w http.ResponseWriter, r *http.Request) {

	var addTagsForLinksReq AddTagsForLinksReq
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()
	if err := body.Decode(&addTagsForLinksReq); err != nil {
		h.logger.Error().Err(err).Msg("Cannot decode addTagsForLinksReq")
		e.ErrorDecodingRequest(w, err)
		return
	}
	if err := h.validator.Struct(addTagsForLinksReq); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Validation error")
		}
		util.JsonResponse(w, "Validation failed", http.StatusBadRequest)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)
	q := sqlc.New(h.db)
	params := sqlc.AddTagsParams{
		AccountID: payload.AccountID,
		Column1:   addTagsForLinksReq.TagNames,
		Column3:   addTagsForLinksReq.LinkIDs,
	}

	link_tag, err := q.AddTags(r.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("cannot add tag")
		e.ErrorInternalServer(w, err)
		return
	}

	response := map[string]interface{}{
		"link_tag": link_tag,
		"result":   "true",
	}

	util.JsonResponse(w, response)
}
