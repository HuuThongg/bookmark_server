package api

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func (a *API) Hello(w http.ResponseWriter, r *http.Request) {
	// Respond with a basic message
	w.Write([]byte("hello\n"))

	// Get the current time from the database
	var currentTime time.Time

	// Create a context for the query (you may already have a context in your app)
	ctx := r.Context() // Use the request's context

	// Execute the query
	err := a.db.QueryRow(ctx, "SELECT NOW()").Scan(&currentTime)
	if err != nil {
		http.Error(w, "Failed to get current time", http.StatusInternalServerError)
		log.Printf("Query failed: %v", err)
		return
	}

	// Write the current time to the response
	w.Write([]byte(fmt.Sprintf("Current time: %v\n", currentTime)))
}
