package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/database"
	"github.com/joho/godotenv"
)

import _ "github.com/lib/pq"

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileServerHits.Add(1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", cfg.fileServerHits.Load())))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileServerHits.Store(0)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "OK")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or flop loading .env")
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	defer db.Close()

	if err != nil {
		log.Fatalf("Database at %s could not be opened: %s", dbURL, err.Error())
	}
	dbQueries := database.New(db)

	cfg := &apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
	}

	const port = "8080"
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("Serving on port %s\n", port)

	mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)

	mux.HandleFunc("GET /admin/metrics", cfg.metricHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)

	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	mux.HandleFunc("POST /api/users", cfg.handleUsersPost)

	log.Fatal(server.ListenAndServe())
}
