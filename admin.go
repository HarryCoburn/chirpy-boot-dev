package main

import (
	"fmt"
	"net/http"
	"os"
)

func (cfg *apiConfig) servMetrics(write http.ResponseWriter, request *http.Request) {
	write.Header().Set("Content-Type", "text/html; charset=utf-8")
	write.WriteHeader(http.StatusOK)
	metricsPage := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())
	fmt.Fprint(write, metricsPage)
}

func (cfg *apiConfig) resetMetrics(write http.ResponseWriter, request *http.Request) {
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		write.Header().Set("Content-Type", "text/plain; charset=utf-8")
		write.WriteHeader(http.StatusForbidden)
		return
	}
	write.Header().Set("Content-Type", "text/plain; charset=utf-8")
	write.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	cfg.dbQueries.DeleteUsers(request.Context())
	fmt.Fprintf(write, "Metrics and Users reset")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
