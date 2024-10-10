package api

import (
	e "bookmark/api/resource/common/err"
	"bookmark/auth"
	"bookmark/db/sqlc"
	"bookmark/util"
	"bookmark/vultr"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gocolly/colly"
	"github.com/jackc/pgx/v5/pgtype"
)

type URLV2 struct {
	URL      string `json:"url" validate:"required"`
	FolderId string `json:"folder_id" validate:"required"`
}

func (h *API) AddLinkV2(w http.ResponseWriter, r *http.Request) {
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

	favicon := resp.Header.Get("content-location")
	log.Printf("Favicon: %v", favicon)
	log.Printf("Favicon fetching took %s", time.Since(faviconFetchStart))

	// Measure URL title and heading fetching time
	titleFetchingStart := time.Now()
	var urlTitle string
	var pageTitle, pageHeading string

	c := colly.NewCollector()

	var ogImage, description, ogImage1 string
	c.OnHTML("meta[property='og:image']", func(e *colly.HTMLElement) {
		ogImage = e.Attr("content")
	})
	c.OnHTML("meta[property='og:image:url']", func(e *colly.HTMLElement) {
		ogImage1 = e.Attr("content")
	})
	c.OnHTML("title", func(e *colly.HTMLElement) {
		urlTitle = e.Text
	})
	c.OnHTML("meta[name='description']", func(e *colly.HTMLElement) {
		description = e.Attr("content")
	})

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		pageHeading = e.Text
	})
	err = c.Visit(urlToOpen)
	if err != nil {
		e.ErrorInternalServer(w, err)
		return
	}
	fmt.Println("description", description)

	if pageTitle != "" {
		if pageHeading != "" {
			if len(pageHeading) > len(pageTitle) {
				urlTitle = pageHeading
			} else {
				urlTitle = pageTitle
			}
		} else {
			urlTitle = pageTitle
		}
	} else {
		if pageHeading != "" {
			urlTitle = pageHeading
		} else {
			urlTitle = req.URL
		}
	}
	if ogImage == "" {
		ogImage = ogImage1
	}
	log.Printf("URL title and heading fetching took %s", time.Since(titleFetchingStart))

	payload := r.Context().Value("payload").(*auth.PayLoad)

	// Measure screenshot fetching time
	screenshotFetchStart := time.Now()
	urlScreenshotChan := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		vultr.UploadLinkThumbnail2(urlScreenshotChan, ogImage, host)
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
		FolderID:      pgtype.Text{Valid: true, String: req.FolderId}, // Ensure this matches the expected type
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
