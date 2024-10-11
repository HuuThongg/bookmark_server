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
		http.Error(w, "Failed to fetch the URL", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	// modifiedHTML := replaceLinks(html, assetHost)

	// w.Header().Set("Content-Type", "text/html")
	// w.Header().Set("Cache-Control", "public, max-age=3600") // Cache HTML for 1 hour
	// w.Header().Set("Expires", time.Now().Add(1*time.Hour).Format(http.TimeFormat))
	//
	// if etag := resp.Header.Get("ETag"); etag != "" {
	// 	w.Header().Set("ETag", etag)
	// }

	contentType := resp.Header.Get("Content-Type")
	fmt.Println("contentType", contentType)
	if strings.Contains(contentType, "text/html") {
		htmlContent := new(strings.Builder)
		io.Copy(htmlContent, resp.Body)
		modifiedHTML := rewriteAssetURLs(htmlContent.String(), target, h.config.DOMAIN)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(modifiedHTML))
	} else {
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
func rewriteAssetURLs(html string, target *url.URL, domain string) string {
	baseUrl := target.Scheme + "://" + target.Host
	assetHost := domain + "/public/proxy?url=" + baseUrl
	// baseTag := fmt.Sprintf(`<base href="%v" target="_blank">`, baseUrl)
	// if strings.Contains(html, "<head>") {
	// 	html = strings.Replace(html, "<head>", "<head>\n\t"+baseTag, 1)
	// }
	html = strings.ReplaceAll(html, `href="/`, `href="`+assetHost+`/`)
	html = strings.ReplaceAll(html, `src="/`, `src="`+assetHost+`/`)
	html = strings.ReplaceAll(html, `srcset="/`, `srcset="`+assetHost+`/`)
	// html = strings.ReplaceAll(html, `imagesrcset="/`, `imagesrcset="`+assetHost+`/`)
	return html
}
