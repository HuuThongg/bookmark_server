package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func Response(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResponse, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(jsonResponse)
}

// func JsonResponse(w http.ResponseWriter, res ...interface{}) {
// 	w.Header().Set("content-type", "application/json")
// 	w.WriteHeader(200)
// 	json.NewEncoder(w).Encode(res)
// }
//
//

func JsonResponse(w http.ResponseWriter, res ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Use a buffer to improve performance
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(res); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	// Write the buffered response
	w.Write(buf.Bytes())
}
