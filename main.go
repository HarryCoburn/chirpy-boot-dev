package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func servHealth(write http.ResponseWriter, request *http.Request) {
	write.Header().Set("Content-Type", "text/plain; charset=utf-8")
	write.WriteHeader(http.StatusOK)
	write.Write([]byte("OK"))
}

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
	write.Header().Set("Content-Type", "text/plain; charset=utf-8")
	write.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(write, "Metrics reset")
}

func chirpValidate(write http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type errReturn struct {
		Error string `json:"error"`
	}

	type validReturn struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respBody := errReturn{
			Error: "Something went wrong",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			write.WriteHeader(500)
			return
		}
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(400)
		write.Write(dat)
		return
	}
	if len(params.Body) > 140 {
		respBody := errReturn{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			write.WriteHeader(500)
			return
		}
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(400)
		write.Write(dat)
		return
	}
	respBody := validReturn{
		Valid: true,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		write.WriteHeader(500)
		return
	}
	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(200)
	write.Write(dat)
}

func main() {
	apiCfg := apiConfig{}
	servMux := http.NewServeMux()
	servMux.HandleFunc("GET /api/healthz/", servHealth)
	servMux.HandleFunc("POST /api/validate_chirp", chirpValidate)
	servMux.HandleFunc("GET /admin/metrics/", apiCfg.servMetrics)
	servMux.HandleFunc("POST /admin/reset/", apiCfg.resetMetrics)

	fileServ := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	servMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServ))
	var server http.Server
	server.Handler = servMux
	server.Addr = ":8080"
	server.ListenAndServe()

}
