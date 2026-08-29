package e2e

import (
	"net/http"
	"os"
	"time"
)

var client = &http.Client{Timeout: 5 * time.Second}

func getAPIURL() string {
	url := os.Getenv("E2E_API_URL")
	if url == "" {
		return "http://localhost:8082"
	}
	return url
}
