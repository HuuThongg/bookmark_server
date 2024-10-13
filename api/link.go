package api

import (
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/util"
	"bookmark/vultr"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	e "bookmark/api/resource/common/err"
)

func (h *API) GetRootLinks(w http.ResponseWriter, r *http.Request) {
	account_id := chi.URLParam(r, "accountID")

	payload := r.Context().Value("payload").(*auth.PayLoad)

	accountID, err := strconv.Atoi(account_id)
	if err != nil {
		log.Println(err)
		util.Response(w, internalServerError, http.StatusInternalServerError)
		return
	}

	if payload.AccountID != int64(accountID) {
		log.Println("account IDs do not match")
		util.Response(w, errors.New("invalid request").Error(), http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)

	links, err := q.GetRootLinks(r.Context(), payload.AccountID)
	if err != nil {
		log.Println(err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	util.JsonResponse(w, links)
}

type renameLinkRequest struct {
	LinkTitle string `json:"link_title"`
	LinkID    string `json:"link_id"`
}

func (r renameLinkRequest) Validate(requestVaidatinChan chan error) error {
	validationError := validation.ValidateStruct(&r,
		validation.Field(&r.LinkTitle, validation.Required.Error("link title is required")),
		validation.Field(&r.LinkID, validation.Required.Error("link id is required"), validation.Length(33, 33).Error("link id must be 33 characters long")),
	)

	requestVaidatinChan <- validationError

	return validationError
}

func (h *API) RenameLink(w http.ResponseWriter, r *http.Request) {

	requestBody := json.NewDecoder(r.Body)

	requestBody.DisallowUnknownFields()

	var req renameLinkRequest

	if err := requestBody.Decode(&req); err != nil {
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

	validationError := <-requestValidationChan
	if validationError != nil {
		e.ErrorInvalidRequest(w, validationError)
		return
	}

	q := sqlc.New(h.db)

	renameLinkParams := sqlc.RenameLinkParams{
		LinkTitle: req.LinkTitle,
		LinkID:    req.LinkID,
	}

	link, err := q.RenameLink(r.Context(), renameLinkParams)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, link)
}

type moveLinksRequest struct {
	Links   []string `json:"links"`
	FolerID string   `json:"folder_id"`
}

func (m moveLinksRequest) Validate(requestValidationChan chan error) error {
	validationError := validation.ValidateStruct(&m,
		validation.Field(&m.Links, validation.Required, validation.Each(validation.Length(33, 33).Error("link id must be 33 characters long"))),
		validation.Field(&m.FolerID, validation.When(m.FolerID != "", validation.Length(33, 33).Error("folder id must be 33 characters long"))),
	)

	requestValidationChan <- validationError

	return validationError
}

func (h *API) MoveLinks(w http.ResponseWriter, r *http.Request) {
	requestBody := json.NewDecoder(r.Body)

	requestBody.DisallowUnknownFields()

	var req moveLinksRequest

	if err := requestBody.Decode(&req); err != nil {
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

	validationError := <-requestValidationChan
	if validationError != nil {
		e.ErrorInvalidRequest(w, validationError)
		return
	}

	q := sqlc.New(h.db)

	if req.FolerID == "" {
		moveLinksToRoot(q, req.Links, w, r.Context())
	} else {
		moveLinksToFolder(q, req.Links, req.FolerID, w, r.Context())
	}
}

func moveLinksToRoot(q *sqlc.Queries, links []string, w http.ResponseWriter, ctx context.Context) {
	var linksMoved []sqlc.Link

	for _, linkID := range links {
		link, err := q.MoveLinkToRoot(ctx, linkID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		linksMoved = append(linksMoved, link)
	}

	util.JsonResponse(w, linksMoved)
}

func moveLinksToFolder(q *sqlc.Queries, links []string, folderID string, w http.ResponseWriter, ctx context.Context) {
	var linksMoved []sqlc.Link

	for _, linkID := range links {
		params := sqlc.MoveLinkToFolderParams{
			FolderID: pgtype.Text{String: folderID, Valid: true},
			LinkID:   linkID,
		}
		link, err := q.MoveLinkToFolder(ctx, params)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		linksMoved = append(linksMoved, link)
	}

	util.JsonResponse(w, linksMoved)
}

type moveLinksToTrashRequest struct {
	LinkIDS []string `json:"link_ids"`
}

func (m moveLinksToTrashRequest) Validate(requestValidationChan chan error) error {
	requestValidationError := validation.ValidateStruct(&m,
		validation.Field(&m.LinkIDS, validation.Required.Error("link id/ids required"), validation.Each(validation.Length(33, 33).Error("link id must be 33 characters long"))),
	)

	requestValidationChan <- requestValidationError

	return requestValidationError
}

func (h *API) MoveLinksToTrash(w http.ResponseWriter, r *http.Request) {
	requestBody := json.NewDecoder(r.Body)

	requestBody.DisallowUnknownFields()

	var req moveLinksToTrashRequest

	if err := requestBody.Decode(&req); err != nil {
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

	validationError := <-requestValidationChan
	if validationError != nil {
		e.ErrorInvalidRequest(w, validationError)
		return
	}

	q := sqlc.New(h.db)

	var trashedLinks []sqlc.Link

	for _, linkID := range req.LinkIDS {
		link, err := q.MoveLinkToTrash(r.Context(), linkID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		trashedLinks = append(trashedLinks, link)
	}

	util.JsonResponse(w, trashedLinks)
}

type restoreLinksRequest struct {
	LinkIDS []string `json:"link_ids"`
}

func (r restoreLinksRequest) Validate(requestValidationChan chan error) error {
	requestValidationChan <- validation.ValidateStruct(&r,
		validation.Field(&r.LinkIDS, validation.Required.When(len(r.LinkIDS) > 0), validation.Each(validation.Length(33, 33).Error("each link id must be 33 characters long"))),
	)
	return validation.ValidateStruct(&r,
		validation.Field(&r.LinkIDS, validation.Required.When(len(r.LinkIDS) > 0), validation.Each(validation.Length(33, 33).Error("each link id must be 33 characters long"))),
	)
}
func (h *API) RestoreLinksFromTrash(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req restoreLinksRequest

	if err := body.Decode(&req); err != nil {
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

	err := <-requestValidationChan
	if err != nil {
		log.Println(err)
		e.ErrorInvalidRequest(w, err)
		return
	}

	q := sqlc.New(h.db)

	var links []sqlc.Link

	for _, linkID := range req.LinkIDS {
		l, err := q.RestoreLinkFromTrash(r.Context(), linkID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		links = append(links, l)
	}

	util.JsonResponse(w, links)
}

type deleteLinksForeverRequest struct {
	LinkIDS []string `json:"link_ids"`
}

func (d deleteLinksForeverRequest) Validate(requestValidationChan chan error) error {
	requestValidationChan <- validation.ValidateStruct(&d,
		validation.Field(&d.LinkIDS, validation.Required.When(len(d.LinkIDS) > 0), validation.Each(validation.Length(33, 33).Error("each link id must be 33 characters long"))),
	)
	return validation.ValidateStruct(&d,
		validation.Field(&d.LinkIDS, validation.Required.When(len(d.LinkIDS) > 0), validation.Each(validation.Length(33, 33).Error("each link id must be 33 characters long"))),
	)
}

func (h *API) DeleteLinksForever(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req deleteLinksForeverRequest

	if err := body.Decode(&req); err != nil {
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

	err := <-requestValidationChan
	if err != nil {
		log.Println(err)
		e.ErrorInvalidRequest(w, err)

		return
	}

	q := sqlc.New(h.db)

	payload := r.Context().Value("payload").(*auth.PayLoad)
	var links []sqlc.Link

	for _, linkID := range req.LinkIDS {
		// get link
		params := sqlc.GetLinkParams{
			LinkID:    linkID,
			AccountID: payload.AccountID,
		}
		link, err := q.GetLink(r.Context(), params)
		if err != nil {
			log.Println("can get link", err)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		tn := strings.Split(link.LinkThumbnail, "/")

		key := (tn[len(tn)-1])

		// log.Println(key)

		// linkScreenshotKey := strings.Split(link.LinkThumbnail, "/")[4]

		// linkFaviconKey := strings.Split(link.LinkFavicon, "/")[4]

		vultr.DeleteObjectFromBucket("/link-thumbnails", key)

		vultr.DeleteObjectFromBucket("/link-favicons", key)

		// if err := util.DeleteFileFromBucket("/screenshots", linkScreenshotKey); err != nil {
		// 	log.Printf("could not delete screenshot from spaces %v", err)
		// 	util.Response(w, "something went wrong", http.StatusInternalServerError)
		// 	return
		// }

		// if err := util.DeleteFileFromBucket("/favicons", linkFaviconKey); err != nil {
		// 	log.Printf("could not delete favicon from spaces %v", err)
		// 	util.Response(w, "something went wrong", http.StatusInternalServerError)
		// 	return
		// }

		l, err := q.DeleteLinkForever(r.Context(), link.LinkID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		links = append(links, l)
	}

	util.JsonResponse(w, links)
}

type TagDto struct {
	TagID   int64  `json:"tag_id"`
	TagName string `json:"tag_name"`
}
type GetFolderLinksRowNew struct {
	LinkID        string             `json:"link_id"`
	LinkTitle     string             `json:"link_title"`
	LinkThumbnail string             `json:"link_thumbnail"`
	LinkFavicon   string             `json:"link_favicon"`
	LinkHostname  string             `json:"link_hostname"`
	LinkUrl       string             `json:"link_url"`
	LinkNotes     string             `json:"link_notes"`
	AccountID     int64              `json:"account_id"`
	FolderID      pgtype.Text        `json:"folder_id"`
	AddedAt       pgtype.Timestamptz `json:"added_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	DeletedAt     pgtype.Timestamptz `json:"deleted_at"`
	Description   pgtype.Text        `json:"description"`
	FolderName    pgtype.Text        `json:"folder_name"`
	Tags          []TagDto           `json:"tags"`
}

func (h *API) GetFolderLinks(w http.ResponseWriter, r *http.Request) {
	accontID, err := strconv.Atoi(chi.URLParam(r, "accountID"))
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	folderID := chi.URLParam(r, "folderID")

	payload := r.Context().Value("payload").(*auth.PayLoad)

	if int64(accontID) != payload.AccountID {
		log.Println("account_id from request not equal to payload account_id")
		util.Response(w, errors.New("account ids do not match").Error(), 404)
		return
	}

	q := sqlc.New(h.db)

	links, err := q.GetFolderLinks(r.Context(), pgtype.Text{String: folderID, Valid: true})
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	var folderLinks []GetFolderLinksRowNew

	// Iterate over the results to unmarshal tags and create FolderLinks
	for _, link := range links {
		var tags []TagDto

		// Unmarshal the tags from []byte to []TagDto
		if len(link.Tags) > 0 {
			err := json.Unmarshal(link.Tags, &tags)
			if err != nil {
				e.ErrorInternalServer(w, err)
				return
			}
		}

		folderLink := GetFolderLinksRowNew{
			LinkID:        link.LinkID,
			LinkTitle:     link.LinkTitle,
			LinkThumbnail: link.LinkThumbnail,
			LinkFavicon:   link.LinkFavicon,
			LinkNotes:     link.LinkNotes,
			LinkHostname:  link.LinkHostname,
			LinkUrl:       link.LinkHostname,
			AccountID:     link.AccountID,
			UpdatedAt:     link.UpdatedAt,
			DeletedAt:     link.DeletedAt,
			Description:   link.Description,
			FolderID:      link.FolderID,
			FolderName:    link.FolderName,
			AddedAt:       link.AddedAt,
			Tags:          tags,
		}

		folderLinks = append(folderLinks, folderLink)
	}
	util.JsonResponse(w, folderLinks)
}

func (h *API) GetAllLinks(w http.ResponseWriter, r *http.Request) {
	accontID, err := strconv.Atoi(chi.URLParam(r, "accountID"))
	if err != nil {
		h.logger.Error().Err(err).Msg("accountID is Empty")
		e.ErrorInternalServer(w, err)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	if int64(accontID) != payload.AccountID {
		log.Println("account_id from request not equal to payload account_id")
		util.Response(w, errors.New("account ids do not match").Error(), 404)
		return
	}

	q := sqlc.New(h.db)

	links, err := q.GetAllLinks(r.Context(), payload.AccountID)
	if err != nil {

		h.logger.Error().Err(err).Msg("can not gett all Link")
		e.ErrorInternalServer(w, err)
		return
	}

	var folderLinks []GetFolderLinksRowNew

	// Iterate over the results to unmarshal tags and create FolderLinks
	for _, link := range links {
		var tags []TagDto

		// Unmarshal the tags from []byte to []TagDto
		if len(link.Tags) > 0 {
			err := json.Unmarshal(link.Tags, &tags)
			if err != nil {
				e.ErrorInternalServer(w, err)
				return
			}
		}

		folderLink := GetFolderLinksRowNew{
			LinkID:        link.LinkID,
			LinkTitle:     link.LinkTitle,
			LinkThumbnail: link.LinkThumbnail,
			LinkFavicon:   link.LinkFavicon,
			LinkNotes:     link.LinkNotes,
			LinkHostname:  link.LinkHostname,
			LinkUrl:       link.LinkHostname,
			AccountID:     link.AccountID,
			UpdatedAt:     link.UpdatedAt,
			DeletedAt:     link.DeletedAt,
			Description:   link.Description,
			FolderID:      link.FolderID,
			FolderName:    link.FolderName,
			AddedAt:       link.AddedAt,
			Tags:          tags,
		}

		folderLinks = append(folderLinks, folderLink)
	}
	util.JsonResponse(w, folderLinks)
}

type AddNoteRequest struct {
	Note   string `json:"note" validate:"required,max=1000"`
	LinkID string `json:"link_id" validate:"required"`
}

func (h *API) AddNote(w http.ResponseWriter, r *http.Request) {

	var addNoteRequest AddNoteRequest

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&addNoteRequest); err != nil {
		h.logger.Error().Err(err).Msg("Cannot decode addNoteRequest")
		e.ErrorDecodingRequest(w, err)
		return
	}

	if err := h.validator.Struct(addNoteRequest); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Validation error")
		}
		util.JsonResponse(w, "Validation failed", http.StatusBadRequest)
		return
	}
	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)
	params := sqlc.AddNoteParams{
		AccountID: payload.AccountID,
		LinkID:    addNoteRequest.LinkID,
		LinkNotes: addNoteRequest.Note,
	}
	linkIdAndNote, err := q.AddNote(r.Context(), params)
	if err != nil {

		h.logger.Error().Err(err).Msg("can not add a note")
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, linkIdAndNote)
}

func (h *API) GetSingleLink(w http.ResponseWriter, r *http.Request) {
	linkID := chi.URLParam(r, "linkID")
	if linkID == "" {
		h.logger.Error().Msg("LinkID is Empty")
		e.ErrorInternalServer(w, errors.New("linkId is empty"))
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)
	params := sqlc.GetLinkParams{
		LinkID:    linkID,
		AccountID: payload.AccountID,
	}

	link, errLink := q.GetLink(r.Context(), params)
	if errLink != nil {
		h.logger.Error().Msg("cannot get link")
		e.ErrorInternalServer(w, errors.New("cannot get link"))
		return

	}
	util.JsonResponse(w, link)
}

type ChangeTitleReq struct {
	LinkTitle string `json:"link_title" validate:"required,max=1000"`
	LinkID    string `json:"link_id" validate:"required"`
}

func (h *API) ChangeLinkTitle(w http.ResponseWriter, r *http.Request) {

	var changeTitleReq ChangeTitleReq

	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()
	if err := body.Decode(&changeTitleReq); err != nil {
		h.logger.Error().Err(err).Msg("Cannot decode ChangeTitleReq")
		e.ErrorDecodingRequest(w, err)
		return
	}

	if err := h.validator.Struct(changeTitleReq); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Validation error")
		}
		util.JsonResponse(w, "Validation failed", http.StatusBadRequest)
		return
	}
	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)
	log.Println("new title", changeTitleReq.LinkTitle)
	params := sqlc.ChangeLinkTitleParams{
		AccountID: payload.AccountID,
		LinkID:    changeTitleReq.LinkID,
		LinkTitle: changeTitleReq.LinkTitle,
	}
	newTitle, err := q.ChangeLinkTitle(r.Context(), params)
	if err != nil {

		h.logger.Error().Err(err).Msg("can notchange link title")
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, newTitle)
}

type ChangeLinkURL struct {
	LinkID  string `json:"link_id" validate:"required,max=1000"`
	LinkURL string `json:"link_URL" validate:"required"`
}

func (h *API) ChangeLinkURL(w http.ResponseWriter, r *http.Request) {

	var changeURLReq ChangeLinkURL

	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()
	if err := body.Decode(&changeURLReq); err != nil {
		h.logger.Error().Err(err).Msg("Cannot decode ChangeLinkURL")
		e.ErrorDecodingRequest(w, err)
		return
	}

	if err := h.validator.Struct(changeURLReq); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Validation error")
		}
		util.JsonResponse(w, "Validation failed", http.StatusBadRequest)
		return
	}
	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)
	params := sqlc.ChangeLinkURLParams{
		AccountID: payload.AccountID,
		LinkID:    changeURLReq.LinkID,
		LinkUrl:   changeURLReq.LinkURL,
	}

	newURL, err := q.ChangeLinkURL(r.Context(), params)
	if err != nil {

		h.logger.Error().Err(err).Msg("can notchange link title")
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, newURL)
}

func (h *API) GetDeletedLinks(w http.ResponseWriter, r *http.Request) {

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	links, err := q.GetAllDeletedLinks(r.Context(), payload.AccountID)
	if err != nil {
		h.logger.Error().Err(err).Msg("cannot get all deleted links")
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, links)
}

type ChangeLinkDesc struct {
	LinkID   string `json:"link_id" validate:"required"`
	LinkDesc string `json:"link_desc" validate:"required"`
}

func (h *API) ChangeLinkDesc(w http.ResponseWriter, r *http.Request) {

	var changeLinkDesc ChangeLinkDesc

	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()
	if err := body.Decode(&changeLinkDesc); err != nil {
		h.logger.Error().Err(err).Msg("Cannot decode ChangeLinkURL")
		e.ErrorDecodingRequest(w, err)
		return
	}

	if err := h.validator.Struct(changeLinkDesc); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Validation error")
		}
		util.JsonResponse(w, "Validation failed", http.StatusBadRequest)
		return
	}
	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)
	params := sqlc.UpdateLinkDescParams{
		AccountID:   payload.AccountID,
		LinkID:      changeLinkDesc.LinkID,
		Description: pgtype.Text{Valid: true, String: changeLinkDesc.LinkDesc},
	}

	newDescription, err := q.UpdateLinkDesc(r.Context(), params)
	if err != nil {

		h.logger.Error().Err(err).Msg("can notchange link title")
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, newDescription)
}
