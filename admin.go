package main

import (
	"fmt"
	"net/http"
	"os"
)

func (cfg *apiConfig) servMetrics(w http.ResponseWriter, r *http.Request) {
	page := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())
	respondWith(w, http.StatusOK, contentTypeHTML, page)
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		respondWith(w, http.StatusForbidden, contentTypePlain, "")
		return
	}

	cfg.fileserverHits.Store(0)
	if err := cfg.dbQueries.DeleteUsers(r.Context()); err != nil {
		respondWith(w, http.StatusInternalServerError, contentTypePlain, "Could not reset users")
		return
	}

	respondWith(w, http.StatusOK, contentTypePlain, "Metrics and Users Reset")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
