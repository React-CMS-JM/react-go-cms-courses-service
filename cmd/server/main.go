package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"react-go-cms-courses-service/internal/courses"
	"react-go-cms-courses-service/internal/platform"
)

func main() {
	cfg := platform.LoadConfig("8083")
	db, err := platform.OpenDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database: %v", err)
	}
	store := courses.NewStore(db)
	h := courses.NewHandler(store, cfg.JWTSecret, cfg.JWTIssuer)
	mux := chi.NewRouter()
	h.Register(mux)
	addr := ":" + cfg.Port
	log.Printf("react-go-cms-courses-service listening on %s", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           platform.CORS(cfg.CORSOrigins, mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
