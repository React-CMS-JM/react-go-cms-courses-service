package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"react-go-cms-courses-service/internal/application/service/course"
	"react-go-cms-courses-service/internal/application/service/lesson"
	"react-go-cms-courses-service/internal/handler"
	"react-go-cms-courses-service/internal/infrastructure/configuration"
	"react-go-cms-courses-service/internal/infrastructure/httpx"
	"react-go-cms-courses-service/internal/infrastructure/repository/repo"
)

const (
	defaultPort         = "8083"
	databasePingTimeout = 10 * time.Second
	readHeaderTimeout   = 10 * time.Second
)

func main() {
	var cfg configuration.Config
	cfg = configuration.LoadConfig(defaultPort)
	var db *sql.DB
	var err error
	db, err = configuration.OpenDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database: %v", err)
	}
	var courseRepository *repo.CourseRepository
	courseRepository = repo.NewCourseRepository(db)
	var lessonRepository *repo.LessonRepository
	lessonRepository = repo.NewLessonRepository(db)
	var courseService *course.Service
	courseService = course.New(courseRepository)
	var lessonService *lesson.Service
	lessonService = lesson.New(lessonRepository)
	var httpHandler *handler.Handler
	httpHandler = handler.New(courseService, lessonService, cfg.JWTSecret, cfg.JWTIssuer)
	var mux *chi.Mux
	mux = chi.NewRouter()
	httpHandler.Register(mux)
	addr := ":" + cfg.Port
	log.Printf("react-go-cms-courses-service listening on %s", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           httpx.CORS(cfg.CORSOrigins, mux),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	log.Fatal(server.ListenAndServe())
}
