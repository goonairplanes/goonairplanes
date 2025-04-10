package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "Hello from Go on Airplanes API route!",
		"timestamp": time.Now().Format(time.RFC3339),
		"method":    r.Method,
		"path":      r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
