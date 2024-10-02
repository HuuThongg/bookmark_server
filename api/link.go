package api

import (
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/util"
	"bookmark/vultr"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
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

type URL struct {
	URL      string `json:"url"`
	FolderId string `json:"folder_id"`
}

func (u URL) validate(requestValidateChan chan error) error {
	validationError := validation.ValidateStruct(&u,
		validation.Field(&u.URL, validation.Required.Error("url is required"), is.URL.Error("url must be a valid url")),
	)

	requestValidateChan <- validationError
	return validationError
}

func (h *API) AddLink(w http.ResponseWriter, r *http.Request) {
	// Measure total request processing time
	start := time.Now()
	defer func() {
		log.Printf("Total processing time: %s", time.Since(start))
	}()

	// Measure validation time
	validationStart := time.Now()
	log.Printf("Validation started")
	rBody := json.NewDecoder(r.Body)
	rBody.DisallowUnknownFields()

	var req URL
	if err := rBody.Decode(&req); err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	requestValidationChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req.validate(requestValidationChan)
	}()
	validationError := <-requestValidationChan

	if validationError != nil {
		e.ErrorInvalidRequest(w, validationError)
		return
	}
	log.Printf("Validation took %s", time.Since(validationStart))

	// Measure URL parsing time
	urlParseStart := time.Now()
	var host string
	var urlToOpen string

	// Parsing URL logic...
	if strings.Contains(req.URL, "?") {
		u, err := url.ParseRequestURI(req.URL)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}
		if u.Scheme == "https" {
			host = u.Host
			urlToOpen = fmt.Sprintf(`%v`, u)
		} else {
			util.Response(w, "invalid url", http.StatusBadRequest)
			return
		}
	} else {
		parsedUrl, err := url.Parse(req.URL)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}

		if parsedUrl.Scheme == "https" {
			host = parsedUrl.Host
			urlToOpen = req.URL
		} else {
			host = parsedUrl.String()
			urlToOpen = fmt.Sprintf(`https://%s`, req.URL)
		}
	}
	log.Printf("URL parsing took %s", time.Since(urlParseStart))

	// Measure favicon fetching time
	faviconFetchStart := time.Now()
	resp, err := http.Get(fmt.Sprintf("https://www.google.com/s2/favicons?domain=%v&sz=64", req.URL))
	if err != nil {
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	log.Printf("Favicon response header: %v", resp.Header.Get("content-location"))
	var favicon string
	if err := util.DownloadFavicon(resp.Header.Get("content-location"), "favicon.icon"); err != nil {
		favicon = resp.Header.Get("content-location")
	}
	log.Printf("Favicon: %v", favicon)
	log.Printf("Favicon fetching took %s", time.Since(faviconFetchStart))

	// Measure browser actions time
	browserActionsStart := time.Now()
	u := launcher.New().UserDataDir("~/.config/google-chrome").Leakless(true).NoSandbox(true).Headless(true).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	page := browser.MustPage(urlToOpen).MustWaitLoad()
	defer browser.MustClose()
	log.Printf("Browser actions took %s", time.Since(browserActionsStart))

	// Measure URL title and heading fetching time
	titleFetchingStart := time.Now()
	var urlTitle string
	urlTiltleChan := make(chan string, 1)
	urlHeadingChan := make(chan string, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		util.GetUrlTitle(page, urlTiltleChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		util.GetUrlHeading(page, urlHeadingChan)
	}()

	title := strings.TrimSpace(<-urlTiltleChan)
	heading := strings.TrimSpace(<-urlHeadingChan)
	log.Printf("URL title and heading fetching took %s", time.Since(titleFetchingStart))

	// Determine the URL title...
	if title != "" {
		if heading != "" {
			if len(heading) > len(title) {
				urlTitle = heading
			} else {
				urlTitle = title
			}
		} else {
			urlTitle = title
		}
	} else {
		if heading != "" {
			urlTitle = heading
		} else {
			urlTitle = req.URL
		}
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	// Measure screenshot fetching time
	screenshotFetchStart := time.Now()
	util.RodGetUrlScreenshot(page)
	urlScreenshotChan := make(chan string, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		vultr.UploadLinkThumbnail(urlScreenshotChan)
	}()
	urlScreenshotLink := <-urlScreenshotChan
	log.Printf("Screenshot fetching took %s", time.Since(screenshotFetchStart))

	// Measure random string generation time
	randomStringGenStart := time.Now()
	stringChan := make(chan string, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		util.RandomStringGenerator(stringChan)
	}()
	linkID := <-stringChan
	log.Printf("Random string generation took %s", time.Since(randomStringGenStart))

	addLinkParams := sqlc.AddLinkParams{
		LinkID:        linkID,
		LinkTitle:     urlTitle,
		LinkHostname:  host,
		LinkUrl:       req.URL,
		LinkFavicon:   favicon,
		AccountID:     payload.AccountID,
		FolderID:      req.FolderId, // Ensure this matches the expected type
		LinkThumbnail: urlScreenshotLink,
	}

	log.Println("AddLink parameters:", addLinkParams)
	q := sqlc.New(h.db)
	link, err := q.AddLink(r.Context(), addLinkParams)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}
	util.JsonResponse(w, link)
	wg.Wait() // Wait for all goroutines to finish
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

	var links []sqlc.Link

	for _, linkID := range req.LinkIDS {
		// get link
		link, err := q.GetLink(r.Context(), linkID)
		if err != nil {
			log.Println(err)
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

	// params := sqlc.GetFolderLinksParams{
	// 	AccountID: payload.AccountID,
	// 	FolderID:  sql.NullString{String: folderID, Valid: true},
	// }

	links, err := q.GetFolderLinks(r.Context(), pgtype.Text{String: folderID, Valid: true})
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}

	util.JsonResponse(w, links)
}
