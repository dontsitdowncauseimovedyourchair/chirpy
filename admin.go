package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) middlewareDevOnlyEndpoint(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.platform != "dev" {
			respondWithError(w, http.StatusForbidden, "Admins pro solamente")
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits.Store(0)
	err := cfg.db.WipeUsers(r.Context())
	if err != nil {
		log.Println("Flop wiping users")
		respondWithError(w, http.StatusInternalServerError, "Flopped on our side. Sorry")
		return
	}
	log.Println("Users have been reset.")
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
