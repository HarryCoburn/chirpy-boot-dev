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
	secret         string
}

func main() {
	// env handling
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("SEC")

	// Open database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Cannot open database: %v", err)
		return
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		dbQueries: dbQueries,
		secret:    secret,
	}

	servMux := http.NewServeMux()
	servMux.HandleFunc("GET /api/healthz/", servHealth)
	servMux.HandleFunc("POST /api/chirps", apiCfg.chirpHandler)
	servMux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	servMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpHandler)
	servMux.HandleFunc("GET /admin/metrics/", apiCfg.servMetrics)
	servMux.HandleFunc("POST /admin/reset/", apiCfg.resetMetrics)
	servMux.HandleFunc("POST /api/users", apiCfg.createNewUser)
	servMux.HandleFunc("POST /api/login", apiCfg.userLogin)
	servMux.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)
	servMux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)

	fileServ := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	servMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServ))

	server := &http.Server{
		Addr:    ":8080",
		Handler: servMux,
	}

	log.Fatal(server.ListenAndServe())
}
