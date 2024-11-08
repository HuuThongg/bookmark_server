package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func LogUserLocationOnDailyRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Get the cookie that stores the last visit date
		value, err := r.Cookie("last_visit_date")
		fmt.Println("last_visit_date", value)
		if err != nil {
			fmt.Println("Expires", err)
			// Get user's IP Address
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			// Fetch location info from IpInfo API
			ip = "135.180.84.155"
			ipInfo, err := getLocationFromIP(ip)
			fmt.Println("id", ip)
			if err != nil {
				log.Printf("Failed to Fetch location data %w", err)
			} else {
				log.Println("hello")
				// log the location info
				logLocation(ipInfo)

				// set a cookie with today's date to prevent logging on subsequent visits
				http.SetCookie(w, &http.Cookie{
					Name:     "last_visit_date",
					Value:    "logged",
					Path:     "/",
					HttpOnly: true,
					Secure:   false,
					SameSite: http.SameSiteLaxMode,
					Expires:  time.Now().Add(12 * time.Hour),
				})
			}

		} else {
			fmt.Fprintf(w, "Cookie found: %s", value.Value)
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func logLocation(ipInfo *IpInfoResponse) {
	logFile, err := os.OpenFile("uesr_location_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		log.Println("Error opening log file:", err)
		return
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Printf("User IP: %s, City: %s, Country: %s\n", ipInfo.Ip, ipInfo.City, ipInfo.Country)
}

type IpInfoResponse struct {
	City     string `json:"city"`
	Country  string `json:"country"`
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
}

func getLocationFromIP(ip string) (*IpInfoResponse, error) {

	apiKey := os.Getenv("IPINFO")
	if apiKey == "" {
		log.Panic("IPINFO is empty")
	}
	fmt.Println("apiKey", apiKey)
	url := fmt.Sprintf("https://ipinfo.io/%s?token=%s", ip, apiKey)
	fmt.Println("url: ", url)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var ipInfo IpInfoResponse
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return nil, err
	}
	fmt.Println("ipInfo", ipInfo)

	return &ipInfo, nil
}
