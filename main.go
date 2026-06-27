package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
