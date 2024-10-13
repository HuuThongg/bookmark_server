package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	e "bookmark/api/resource/common/err"
	"bookmark/auth"

	"bookmark/db/sqlc"

	"bookmark/middleware"

	"bookmark/util"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *API) CreateFolder(w http.ResponseWriter, r *http.Request) {
	// log.Printf("authorized payload: %v", r.Context().Value("authorizedPayload").(*auth.PayLoad))
	requestBody := r.Context().Value("createFolderRequest").(*middleware.CreateFolderRequestBody).Body

	authorizedPayload := r.Context().Value("createFolderRequest").(*middleware.CreateFolderRequestBody).PayLoad

	queries := sqlc.New(h.db)
	if requestBody.FolderID != "null" {
		util.CreateChildFolder(queries, w, r, requestBody.FolderName, requestBody.FolderID, authorizedPayload.AccountID)
		return
	}
	stringChan := make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		util.RandomStringGenerator(stringChan)
	}()

	folderLabelChan := make(chan string, 1)

	wg.Add(1)

	go func() {
		defer wg.Done()

		util.GenFolderLabel(folderLabelChan)
	}()

	folderID := <-stringChan

	folderLabel := <-folderLabelChan

	folderParams := sqlc.CreateFolderParams{
		FolderID:    folderID,
		FolderName:  requestBody.FolderName,
		SubfolderOf: pgtype.Text{Valid: false},
		AccountID:   authorizedPayload.AccountID,
		Path:        folderLabel,
		Label:       folderLabel,
	}
	folder, err := queries.CreateFolder(r.Context(), folderParams)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			h.logger.Error().Msgf("createFolder: %v", err)
			fmt.Println("hey")
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}

	}

	rf := newReturnedFolder(folder)

	util.JsonResponse(w, rf)

	wg.Wait()
}

// CREATE CHILD FOLDER
type createChildFolderRequest struct {
	FolderName   string `json:"folder_name"`
	ParentFolder string `json:"parent_folder"`
}

func (s createChildFolderRequest) validate(reqValidationChan chan error) error {
	returnVal := validation.ValidateStruct(&s,
		validation.Field(&s.FolderName, validation.Required.When(s.FolderName == "").Error("Folder name is required"), validation.Length(1, 1000).Error("Folder name must be at least 1 character long"), validation.Match(regexp.MustCompile("^[^?[\\]{}|\\\\`./!@#$%^&*()_-]+$")).Error("Folder name must not have special characters")),
		validation.Field(&s.ParentFolder, validation.Required.When(s.ParentFolder == "").Error("Parent folder id is required"), validation.Length(33, 33).Error("Parent folder id must be 33 characters long")),
	)
	reqValidationChan <- returnVal

	return returnVal
}

func (h *API) CreateChildFolder(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)

	rBody.DisallowUnknownFields()

	var req createChildFolderRequest

	err := rBody.Decode(&req)
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

	requestValidationErr := <-reqValidationChan
	if requestValidationErr != nil {
		e.ErrorInvalidRequest(w, requestValidationErr)
		return
	}

	wg.Wait()

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	parentFolder, err := q.GetFolder(r.Context(), req.ParentFolder)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	parentFolderPath := parentFolder.Path

	stringChan := make(chan string, 1)

	wg.Add(1)

	go func() {
		defer wg.Done()

		util.RandomStringGenerator(stringChan)
	}()

	folderLabel := <-stringChan

	folderID := <-stringChan

	path := strings.Join([]string{parentFolderPath, folderLabel}, ".")

	arg := sqlc.CreateFolderParams{
		FolderID:    folderID,
		FolderName:  req.FolderName,
		SubfolderOf: pgtype.Text{String: req.ParentFolder, Valid: true},
		AccountID:   payload.AccountID,
		Path:        path,
		Label:       folderLabel,
	}

	createdChildFolder, err := q.CreateFolder(r.Context(), arg)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	wg.Wait()

	util.JsonResponse(w, createdChildFolder)
}

func (h *API) GetRootFolders(w http.ResponseWriter, r *http.Request) {
	fmt.Println("fucku")
	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	folders, err := q.GetRootNodes(r.Context(), payload.AccountID)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	log.Println(folders)

	util.JsonResponse(w, folders)
}

func (h *API) GetFolderChildren(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting folder children...")

	account_id := chi.URLParam(r, "accountID")

	folderID := chi.URLParam(r, "folderID")

	payload := r.Context().Value("payload").(*auth.PayLoad)

	accountID, err := strconv.Atoi(account_id)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	if payload.AccountID != int64(accountID) {
		log.Println("unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(r.Context(), folderID)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	if folder.AccountID != payload.AccountID {
		log.Println("unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if folder.AccountID != int64(accountID) {
		log.Println("unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	childrenFolders, err := q.GetFolderNodes(r.Context(), pgtype.Text{String: folderID, Valid: true})
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, childrenFolders)
}

// GET FOLDER ANCESTORS
func (h *API) GetFolderAncestors(w http.ResponseWriter, r *http.Request) {
	folderID := chi.URLParam(r, "folderID")

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(r.Context(), folderID)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	label := folder.Label

	ancestors, err := q.GetFolderAncestors(r.Context(), label)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, ancestors)
}

// STAR FOLDER
type starFoldersReq struct {
	FolderIDs []string `json:"folder_ids"`
}

func (s starFoldersReq) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&s,
		validation.Field(&s.FolderIDs, validation.Each(validation.Length(33, 33)), validation.Required),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *API) StarFolders(w http.ResponseWriter, r *http.Request) {
	// get and validate folder id
	reqBody := json.NewDecoder(r.Body)

	reqBody.DisallowUnknownFields()

	var req starFoldersReq

	if err := reqBody.Decode(&req); err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		req.Validate(reqValidationChan)
	}()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		e.ErrorInvalidRequest(w, reqValidationErr)
		return
	}

	// get folder
	q := sqlc.New(h.db)

	var starredFolders []sqlc.Folder

	for _, fid := range req.FolderIDs {
		folder, err := q.GetFolder(r.Context(), fid)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		// check if folder belongs to caller
		payload := r.Context().Value("payload").(*auth.PayLoad)

		if folder.AccountID != payload.AccountID {
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// toggle folder star status
		starredFolder, err := q.StarFolder(r.Context(), folder.FolderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		starredFolders = append(starredFolders, starredFolder)
	}

	// wait for go routines to finish
	wg.Wait()

	// return starred of folders
	util.JsonResponse(w, starredFolders)
}

// UNSTAR FOLDERS
type unStarFoldersReq struct {
	FolderIDs []string `json:"folder_ids"`
}

func (s unStarFoldersReq) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&s,
		validation.Field(&s.FolderIDs, validation.Each(validation.Length(33, 33)), validation.Required),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *API) UnstarFolders(w http.ResponseWriter, r *http.Request) {
	reqBody := json.NewDecoder(r.Body)

	reqBody.DisallowUnknownFields()

	var req unStarFoldersReq

	if err := reqBody.Decode(&req); err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		req.Validate(reqValidationChan)
	}()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		e.ErrorInvalidRequest(w, reqValidationErr)
		return
	}

	wg.Wait()

	q := sqlc.New(h.db)

	var unstarredFolders []sqlc.Folder

	for _, folderID := range req.FolderIDs {
		// check if each folder exists
		folder, err := q.GetFolder(r.Context(), folderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		// check if user is authorized ie is the owner of the folders
		payload := r.Context().Value("payload").(*auth.PayLoad)

		if payload.AccountID != folder.AccountID {
			log.Println("unauthorized")
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// unstar each folder
		unstarredFolder, err := q.UnstarFolder(r.Context(), folder.FolderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		unstarredFolders = append(unstarredFolders, unstarredFolder)
	}

	// return unstarred folders
	util.JsonResponse(w, unstarredFolders)
}

// TOGGLE FOLDER STARRED
type toggleFolderStarredReq struct {
	FolderIDs []string `json:"folder_ids"`
}

func (t toggleFolderStarredReq) Validate(rValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&t,
		validation.Field(&t.FolderIDs, validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long")), validation.Required.Error("folder id/ids required")),
	)

	rValidationChan <- validationErr

	return validationErr
}

func (h *API) ToggleFolderStarred(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)

	rBody.DisallowUnknownFields()

	var req toggleFolderStarredReq

	if err := rBody.Decode(&req); err != nil {
		e.ErrorInvalidRequest(w, err)
		return
	}

	rValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(rValidationChan)
	}()

	wg.Wait()

	if err := <-rValidationChan; err != nil {
		e.ErrorInvalidRequest(w, err)
		return
	}

	q := sqlc.New(h.db)

	var foldersStarred []sqlc.Folder

	for _, folderID := range req.FolderIDs {
		folderStarred, err := q.ToggleFolderStarred(r.Context(), folderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		foldersStarred = append(foldersStarred, folderStarred)
	}

	util.JsonResponse(w, foldersStarred)
}

// RENAME FOLDER
type renameFolder struct {
	NewFolderName string `json:"new_folder_name"`
	FolderID      string `json:"folder_id"`
}

func (s renameFolder) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&s,
		validation.Field(&s.NewFolderName, validation.Required.When(s.NewFolderName == "").Error("New folder name cannot be empty!"), validation.Length(1, 200).Error("Folder name must be atleast 1 character long"), validation.Match(regexp.MustCompile("^[^?[\\]{}|\\\\`./!@$%^&*()_]+$")).Error("Folder name must not have special characters")),
		validation.Field(&s.FolderID, validation.Required.When(s.FolderID == "").Error("Folder id is required"), validation.Length(33, 33).Error("Folder ID must be 33 characters long")),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *API) RenameFolder(w http.ResponseWriter, r *http.Request) {
	reqBody := json.NewDecoder(r.Body)

	reqBody.DisallowUnknownFields()

	var req renameFolder

	if err := reqBody.Decode(&req); err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(reqValidationChan)
	}()

	wg.Wait()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		e.ErrorInvalidRequest(w, reqValidationErr)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(r.Context(), req.FolderID)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	if folder.AccountID != payload.AccountID {
		log.Println("user is unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	arg := sqlc.RenameFolderParams{
		FolderName: req.NewFolderName,
		FolderID:   req.FolderID,
	}

	renamedFolder, err := q.RenameFolder(r.Context(), arg)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, newReturnedFolder(renamedFolder))
}

type moveFoldersToTrash struct {
	FolderIDs []string `json:"folder_ids"`
}

func (s moveFoldersToTrash) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&s,
		validation.Field(&s.FolderIDs, validation.Required.Error("folder ids requiured"), validation.Each(validation.Length(33, 33).Error("folder id must be 33 characters long"))),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *API) MoveFoldersToTrash(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)

	rBody.DisallowUnknownFields()

	var req moveFoldersToTrash

	if err := rBody.Decode(&req); err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		req.Validate(reqValidationChan)
	}()

	wg.Wait()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		e.ErrorInvalidRequest(w, reqValidationErr)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	var trashedFolders []sqlc.Folder

	for _, folderID := range req.FolderIDs {
		folder, err := q.GetFolder(r.Context(), folderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		if folder.AccountID != payload.AccountID {
			log.Println("user unauthorized to delete this folder")
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		trashedFolder, err := q.MoveFolderToTrash(r.Context(), folder.FolderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		trashedFolders = append(trashedFolders, trashedFolder)
	}

	util.JsonResponse(w, trashedFolders)
}

func (h *API) GetFolder(w http.ResponseWriter, r *http.Request) {
	// folderID := chi.URLParam(r, "folderID")
	// payload := r.Context().Value("payload").(*auth.PayLoad)

	body := r.Context().Value("readRequestOnCollectionDetails").(*middleware.ReadRequestOnCollectionDetails)

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(r.Context(), body.FolderID)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	// if folder.AccountID != payload.AccountID {
	// 	log.Println("user unauthorized for this operation")
	// 	util.Response(w, "unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	util.JsonResponse(w, folder)
}

// MOVE FOLDERS
type moveFoldersRequest struct {
	DestinationFolderID string   `json:"destination_folder_id"`
	FolderIDs           []string `json:"folder_ids"`
}

func (m moveFoldersRequest) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&m,
		validation.Field(&m.FolderIDs, validation.Required.Error("Folder IDs requiured"), validation.Each(validation.Length(33, 33).Error("Folder id must be 33 characters long"))),
		validation.Field(&m.DestinationFolderID, validation.Required.Error("Destination folder id required"), validation.Length(33, 33).Error("Destination folder id must be 33 charecters long")),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *API) MoveFolders(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)
	rBody.DisallowUnknownFields()

	var req moveFoldersRequest

	err := rBody.Decode(&req)
	if err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		req.Validate(reqValidationChan)
	}()

	wg.Wait()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		e.ErrorInvalidRequest(w, err)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	destinationFolder, err := q.GetFolder(r.Context(), req.DestinationFolderID)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	if destinationFolder.AccountID != payload.AccountID {
		log.Println("unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var foldersMoved []sqlc.Folder

	for _, folder_ID := range req.FolderIDs {
		folder, err := q.GetFolder(r.Context(), folder_ID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		if folder.AccountID != payload.AccountID {
			log.Println("unauthorized")
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		arg := sqlc.MoveFolderParams{
			Label:   destinationFolder.Label,
			Label_2: folder.Label,
			Label_3: folder.Label,
		}

		movedFolders, err := q.MoveFolder(r.Context(), arg)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		arg2 := sqlc.UpdateFolderSubfolderOfParams{
			SubfolderOf: pgtype.Text{String: destinationFolder.FolderID, Valid: true},
			FolderID:    folder.FolderID,
		}

		_, err = q.UpdateFolderSubfolderOf(r.Context(), arg2)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		for _, movedFolder := range movedFolders {
			movedFolder, err = q.GetFolder(r.Context(), movedFolder.FolderID)
			if err != nil {
				e.ErrorInternalServer(w, err)
				return
			}

			foldersMoved = append(foldersMoved, movedFolder)
		}
	}

	util.JsonResponse(w, foldersMoved)
}

// MOVE FOLDERS TO ROOT
type moveFoldersToRootRequest struct {
	FolderIDs []string `json:"folder_ids"`
}

func (m moveFoldersToRootRequest) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&m,
		validation.Field(&m.FolderIDs, validation.Required.Error("Folder IDs requiured"), validation.Each(validation.Length(33, 33).Error("Folder id must be 33 characters long"))),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *API) MoveFoldersToRoot(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)
	rBody.DisallowUnknownFields()

	var req moveFoldersToRootRequest

	err := rBody.Decode(&req)
	if err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(reqValidationChan)
	}()

	wg.Wait()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		e.ErrorInvalidRequest(w, reqValidationErr)
		return
	}

	q := sqlc.New(h.db)

	payload := r.Context().Value("payload").(*auth.PayLoad)

	var foldersMovedToRoot []sqlc.Folder

	for _, folderID := range req.FolderIDs {
		folder, err := q.GetFolder(r.Context(), folderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		if folder.AccountID != payload.AccountID {
			log.Println("unauthorized")
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		arg := sqlc.MoveFoldersToRootParams{
			Label:   folder.Label,
			Label_2: folder.Label,
		}

		folderMovedToRoot, err := q.MoveFoldersToRoot(r.Context(), arg)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		if err := q.UpdateParentFolderToNull(r.Context(), folder.FolderID); err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		foldersMovedToRoot = append(foldersMovedToRoot, folderMovedToRoot...)
	}

	util.JsonResponse(w, foldersMovedToRoot)
}

type restoreFoldersRequest struct {
	FolderIDS []string `json:"folder_ids"`
}

func (r restoreFoldersRequest) Validate(requestValidationChan chan error) error {
	requestValidationChan <- validation.ValidateStruct(&r,
		validation.Field(&r.FolderIDS, validation.Required.When(len(r.FolderIDS) > 0), validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long"))),
	)
	return validation.ValidateStruct(&r,
		validation.Field(&r.FolderIDS, validation.Required.When(len(r.FolderIDS) > 0), validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long"))),
	)
}

func (h *API) RestoreFoldersFromTrash(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req restoreFoldersRequest

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

	var folders []sqlc.Folder

	for _, folderID := range req.FolderIDS {
		f, err := q.RestoreFolderFromTrash(r.Context(), folderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		folders = append(folders, f)
	}

	util.JsonResponse(w, folders)
}

type deleteFoldersForeverRequest struct {
	FolderIDS []string `json:"folder_ids"`
}

func (d deleteFoldersForeverRequest) Validate(requestValidationChan chan error) error {
	requestValidationChan <- validation.ValidateStruct(&d,
		validation.Field(&d.FolderIDS, validation.Required.When(len(d.FolderIDS) > 0), validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long"))),
	)
	return validation.ValidateStruct(&d,
		validation.Field(&d.FolderIDS, validation.Required.When(len(d.FolderIDS) > 0), validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long"))),
	)
}

func (h *API) DeleteFoldersForever(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req deleteFoldersForeverRequest

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

	var folderIds []string

	for _, folderID := range req.FolderIDS {
		folder_id, err := q.DeleteFolderForever(r.Context(), folderID)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		folderIds = append(folderIds, folder_id)
	}
	response := map[string]interface{}{
		"result":    true,
		"folderIds": folderIds,
	}
	util.JsonResponse(w, response)
}

func (h *API) GetSortedTreeFolders(w http.ResponseWriter, r *http.Request) {

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	folders, err := q.GetTreeFolders(r.Context(), payload.AccountID)
	if err != nil {
		h.logger.Error().Err(err).Msg("cannot get GetTreeFolders")
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, folders)
}

type Folder struct {
	FolderID    string `json:"folder_id" validate:"required"`
	FolderOrder int8   `json:"folder_order" validate:"required,min=0"`
	SubfolderOf string `json:"subfolder_of"`
}

type UpdateFolderSortRequest struct {
	Folders []Folder `json:"folders" validate:"required,dive"`
}

func (h *API) UpdateFolderSort(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)
	body.DisallowUnknownFields()

	var req UpdateFolderSortRequest

	// Decode the incoming request body
	if err := body.Decode(&req); err != nil {
		fmt.Println("can not decod")
		e.ErrorDecodingRequest(w, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Validation error")
		}
		util.JsonResponse(w, "Validation failed", http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)
	payload := r.Context().Value("payload").(*auth.PayLoad)
	for _, folder := range req.Folders {
		params := sqlc.UpdateFolderOrderParams{
			FolderOrder: int32(folder.FolderOrder),
			AccountID:   payload.AccountID,
			FolderID:    folder.FolderID,
		}

		if err := q.UpdateFolderOrder(r.Context(), params); err != nil {
			h.logger.Error().Err(err).Msg("cannot update folder order")
			e.ErrorInternalServer(w, err)
			return
		}
	}

	util.JsonResponse(w, map[string]string{"result": "true"})
}
