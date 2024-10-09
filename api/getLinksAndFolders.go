package api

import (
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/middleware"
	"bookmark/util"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	e "bookmark/api/resource/common/err"

	"github.com/jackc/pgx/v5/pgtype"
)

type getLinksAndFoldersResponse struct {
	Folders []returnFolder `json:"folders"`
	Links   []sqlc.Link    `json:"links"`
}

func newResponse(folders []returnFolder, links []sqlc.Link) *getLinksAndFoldersResponse {
	return &getLinksAndFoldersResponse{
		Folders: folders,
		Links:   links,
	}
}

type returnFolder struct {
	FolderID        string             `json:"folder_id"`
	AccountID       int64              `json:"account_id"`
	FolderName      string             `json:"folder_name"`
	Path            string             `json:"path"`
	Label           string             `json:"label"`
	Starred         bool               `json:"starred"`
	FolderCreatedAt string             `json:"folder_created_at"`
	FolderUpdatedAt string             `json:"folder_updated_at"`
	SubfolderOf     pgtype.Text        `json:"subfolder_of"`
	FolderDeletedAt pgtype.Timestamptz `json:"folder_deleted_at"`
	FolderSort      int32              `json:"folder_sort"`
}

func newReturnedFolder(f sqlc.Folder) returnFolder {
	return returnFolder{
		FolderID:        f.FolderID,
		AccountID:       f.AccountID,
		FolderName:      f.FolderName,
		Path:            f.Path,
		Label:           f.Label,
		Starred:         f.Starred,
		FolderCreatedAt: strings.Join(strings.Split(strings.Split(f.FolderUpdatedAt.Time.Local().Format(time.RFC3339), "T")[0], "-"), "/"),
		FolderUpdatedAt: strings.Join(strings.Split(strings.Split(f.FolderCreatedAt.Time.Local().Format(time.RFC3339), "T")[0], "-"), "/"),
		SubfolderOf:     f.SubfolderOf,
		FolderDeletedAt: f.FolderDeletedAt,
		FolderSort:      f.FolderOrder,
	}
}

func (h *API) GetLinksAndFolders(w http.ResponseWriter, r *http.Request) {
	log.Println("hello get links and folder ")
	// body := r.Context().Value("readRequestOnCollectionDetails")
	body := r.Context().Value("readRequestOnCollectionDetails").(*middleware.ReadRequestOnCollectionDetails)
	if body.FolderID == "null" {
		log.Println("folder id existss")
		fmt.Println("folder id existss1")
		getRootNodesAndLinks(h, body.Payload.AccountID, w, r.Context())
	} else {

		log.Println("folder id not  existss")
		fmt.Println("folder id not existss1")
		getFolderNodesAndLinks(h, body.FolderID, w, r.Context())
	}
}

func (h *API) GetAllFolders(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(*auth.PayLoad)
	q := sqlc.New(h.db)
	fs, err := q.GetFoldersByAccountId(r.Context(), payload.AccountID)
	if err != nil {
		log.Println(err)
		util.Response(w, err.Error(), 500)
		return
	}
	var rfs []returnFolder
	for _, f := range fs {
		folder := newReturnedFolder(f)
		rfs = append(rfs, folder)
	}
	util.JsonResponse(w, rfs)
}

func getRootNodesAndLinks(h *API, accountID int64, w http.ResponseWriter, ctx context.Context) {
	q := sqlc.New(h.db)
	fs, err := q.GetRootFolders(ctx, accountID)
	if err != nil {
		log.Println(err)
		util.Response(w, err.Error(), 500)
		return
	}
	var rfs []returnFolder
	for _, f := range fs {
		folder := newReturnedFolder(f)
		rfs = append(rfs, folder)
	}

	links, err := q.GetRootLinks(ctx, accountID)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}
	res := newResponse(rfs, links)
	util.JsonResponse(w, res)
}
func getFolderNodesAndLinks(h *API, folderID string, w http.ResponseWriter, ctx context.Context) {
	q := sqlc.New(h.db)

	nodes, err := q.GetFolderNodes(ctx, pgtype.Text{String: folderID, Valid: true})
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}
	var rfs []returnFolder

	for _, n := range nodes {
		f := newReturnedFolder(n)
		rfs = append(rfs, f)
	}

	links, err := q.GetFolderLinks(ctx, pgtype.Text{String: folderID, Valid: true})
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}
	res := newResponse(rfs, links)
	util.JsonResponse(w, res)
}
