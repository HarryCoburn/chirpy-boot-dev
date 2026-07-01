package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/HarryCoburn/chirpy-boot-dev/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

func main() {
	// env handling
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	// Open database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Cannot open databse: %v", err)
		return
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{}
	apiCfg.dbQueries = dbQueries

	servMux := http.NewServeMux()
	servMux.HandleFunc("GET /api/healthz/", servHealth)
	servMux.HandleFunc("POST /api/chirps", apiCfg.chirpHandler)
	servMux.HandleFunc("GET /admin/metrics/", apiCfg.servMetrics)
	servMux.HandleFunc("POST /admin/reset/", apiCfg.resetMetrics)
	servMux.HandleFunc("POST /api/users", apiCfg.createNewUser)

	fileServ := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	servMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServ))
	var server http.Server
	server.Handler = servMux
	server.Addr = ":8080"
	server.ListenAndServe()
}
