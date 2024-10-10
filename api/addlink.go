package api

import (
	e "bookmark/api/resource/common/err"
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/util"
	"bookmark/vultr"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gocolly/colly"
	"github.com/jackc/pgx/v5/pgtype"
)

type PageData struct {
	Title       string
	Description string
	OgImage     string
	Error       string
}
type URLV2 struct {
	URL      string `json:"url" validate:"required"`
	FolderId string `json:"folder_id" validate:"required"`
}

func (h *API) AddLinkV2(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		log.Printf("Total processing time: %s", time.Since(start))
	}()

	validationStart := time.Now()
	log.Printf("Validation started")
	rBody := json.NewDecoder(r.Body)
	rBody.DisallowUnknownFields()

	var req URLV2
	if err := rBody.Decode(&req); err != nil {
		e.ErrorDecodingRequest(w, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			h.logger.Error().Err(err).Msg("Can not decode URLV2")
			e.ErrorInternalServer(w, err)
			return
		}
	}

	log.Printf("Validation took %s", time.Since(validationStart))

	// Measure random string generation time
	randomStringGenStart := time.Now()
	stringChan := make(chan string, 1)
	go func() {
		util.RandomStringGenerator(stringChan)
	}()
	linkID := <-stringChan
	log.Printf("Random string generation took %s", time.Since(randomStringGenStart))

	urlParseStart := time.Now()
	host, urlToOpen, onlyStoreURL := processUrl(req.URL)

	q := sqlc.New(h.db)

	payload := r.Context().Value("payload").(*auth.PayLoad)

	if onlyStoreURL {
		addLinkParams := sqlc.AddLinkParams{
			LinkID:        linkID,
			LinkTitle:     "",
			LinkHostname:  host,
			LinkUrl:       req.URL,
			LinkFavicon:   "",
			AccountID:     payload.AccountID,
			FolderID:      pgtype.Text{Valid: true, String: req.FolderId}, // Ensure this matches the expected type
			LinkThumbnail: "",
			Description:   pgtype.Text{Valid: false, String: ""},
		}

		link, err := q.AddLink(r.Context(), addLinkParams)
		if err != nil {
			e.ErrorInternalServer(w, err)
			return
		}
		util.JsonResponse(w, link)
	}
	log.Printf("URL parsing took %s", time.Since(urlParseStart))

	faviconFetchStart := time.Now()
	faviconChan := make(chan string, 1)
	ctxFav, cancelFav := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFav()
	go fetchFavicon(ctxFav, req.URL, faviconChan)
	favicon := <-faviconChan

	log.Printf("Favicon: %v", favicon)
	log.Printf("Favicon fetching took %s", time.Since(faviconFetchStart))

	// Measure URL title and heading fetching time
	titleFetchingStart := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	dataChan := make(chan PageData)
	go fetchPageData(ctx, urlToOpen, dataChan)
	pageDataResult := <-dataChan

	log.Printf("URL title and heading fetching took %s", time.Since(titleFetchingStart))

	screenshotFetchStart := time.Now()
	urlScreenshotChan := make(chan string, 1)
	go func() {
		defer close(urlScreenshotChan)
		if err := vultr.UploadLinkThumbnail2(urlScreenshotChan, pageDataResult.OgImage, host); err != nil {
			log.Println("Error:", err)
		}
	}()
	urlScreenshotLink := <-urlScreenshotChan
	log.Printf("Screenshot fetching took %s", time.Since(screenshotFetchStart))

	addLinkParams := sqlc.AddLinkParams{
		LinkID:        linkID,
		LinkTitle:     pageDataResult.Title,
		LinkHostname:  host,
		LinkUrl:       req.URL,
		LinkFavicon:   favicon,
		AccountID:     payload.AccountID,
		FolderID:      pgtype.Text{Valid: true, String: req.FolderId}, // Ensure this matches the expected type
		LinkThumbnail: urlScreenshotLink,
		Description:   pgtype.Text{Valid: true, String: pageDataResult.Description},
	}

	link, err := q.AddLink(r.Context(), addLinkParams)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}
	util.JsonResponse(w, link)
}

func fetchFavicon(ctx context.Context, url string, faviconChan chan<- string) {
	defer close(faviconChan)
	resp, err := http.Get(fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=64", url))
	if err == nil {
		faviconChan <- resp.Header.Get("content-location")
	} else {
		faviconChan <- ""
	}
}

func processUrl(URL string) (host, urlToOpen string, onlyStoreURL bool) {
	onlyStoreURL = false
	u, err := url.ParseRequestURI(URL)
	if err != nil {
		URL = "https://" + URL
		fmt.Println("err")
		u, err = url.ParseRequestURI(URL)
		if err != nil {
			onlyStoreURL = true
			return "", "", onlyStoreURL
		}
		if strings.HasPrefix(URL, "www.") {
			URL = "https://" + URL
			u, err = url.ParseRequestURI(URL)
		}
		if err != nil {
			onlyStoreURL = true
			return "", "", onlyStoreURL
		}
	}
	fmt.Println("dasda")
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	host = u.Host
	urlToOpen = u.String()

	if u.Scheme == "http" {
		urlToOpen = fmt.Sprintf("https://%s", host)
	}

	return host, urlToOpen, onlyStoreURL
}

func fetchPageData(ctx context.Context, url string, dataChan chan<- PageData) {
	defer close(dataChan)

	data := PageData{}

	c := colly.NewCollector()
	// Set up the callbacks for title, description, and OG images
	c.OnHTML("title", func(e *colly.HTMLElement) {
		data.Title = e.Text
	})

	c.OnHTML("meta[name='description']", func(e *colly.HTMLElement) {
		data.Description = e.Attr("content")
	})

	c.OnHTML("meta[property='og:image']", func(e *colly.HTMLElement) {
		data.OgImage = e.Attr("content")
	})

	c.OnHTML("meta[property='og:image:url']", func(e *colly.HTMLElement) {
		if data.OgImage == "" {
			data.OgImage = e.Attr("content")
		}
	})

	// Attempt to visit the URL
	if err := c.Visit(url); err != nil {
		dataChan <- PageData{Error: "failed to visit URL: " + err.Error()} // Send error message
		return
	}

	c.Wait()

	select {
	case dataChan <- data:
	case <-ctx.Done():
		log.Println("Fetch page data canceled")
		dataChan <- PageData{Title: "", Description: "", OgImage: ""}
	}
}
