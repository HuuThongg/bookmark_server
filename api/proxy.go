package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (h *API) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "url query parameter is missing", http.StatusBadRequest)
		return
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	resp, err := client.Get(target.String())
	if err != nil {
		http.Error(w, "Failed to fetch the URL: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")

	if strings.Contains(contentType, "text/html") {
		htmlContent := new(strings.Builder)
		if _, err := io.Copy(htmlContent, resp.Body); err != nil {
			http.Error(w, "Failed to read the response body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		modifiedHTML := rewriteAssetURLs(htmlContent.String(), target, h.config.DOMAIN)
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(modifiedHTML)); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	} else {
		if _, err := io.Copy(w, resp.Body); err != nil {
			http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// Rewrite function that doesn't modify asset links
func rewriteAssetURLs(html string, target *url.URL, domain string) string {
	// Keep the original hrefs as is, just rewrite any necessary script or link tags for your proxy
	fmt.Println("target schmea", target.Scheme)
	html = strings.ReplaceAll(html, `href="`, `href="`+target.Scheme+`://`+target.Host+`/`)
	html = strings.ReplaceAll(html, `src="`, `src="`+target.Scheme+`://`+target.Host+`/`)
	return html
}

// baseTag := fmt.Sprintf(`<base href="%v" target="_blank">`, baseUrl)
// if strings.Contains(html, "<head>") {
// 	html = strings.Replace(html, "<head>", "<head>\n\t"+baseTag, 1)
// }
// modifiedHTML := replaceLinks(html, assetHost)

// html = strings.ReplaceAll(html, `imagesrcset="/`, `imagesrcset="`+assetHost+`/`)
// w.Header().Set("Content-Type", "text/html")
// w.Header().Set("Cache-Control", "public, max-age=3600") // Cache HTML for 1 hour
// w.Header().Set("Expires", time.Now().Add(1*time.Hour).Format(http.TimeFormat))
//
// if etag := resp.Header.Get("ETag"); etag != "" {
// 	w.Header().Set("ETag", etag)
// }
