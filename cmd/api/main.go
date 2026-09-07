package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/teddinator/Internship-tracker/internal/applications"
	"github.com/teddinator/Internship-tracker/internal/companies"
	"github.com/teddinator/Internship-tracker/internal/contacts"
	"github.com/teddinator/Internship-tracker/internal/followups"
	"github.com/teddinator/Internship-tracker/internal/notes"
)

type applicationServer struct {
	db *sql.DB
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.Println("closing database connection")

		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	app := &applicationServer{
		db: db,
	}

	applicationHandler := applications.NewHandler(db)
	companiesHandler := companies.NewHandler(db)
	notesHandler := notes.NewHandler(db)
	contactsHandler := contacts.NewHandler(db)
	followupsHandler := followups.NewHandler(db)

	router := chi.NewRouter()

	router.Get("/health", app.healthHandler)
	router.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Write([]byte("done"))
	})

	router.Route("/applications", func(r chi.Router) {
		r.Get("/", applicationHandler.GetAll)
		r.Get("/export.csv", applicationHandler.ExportCSV)
		r.Post("/", applicationHandler.CreateApp)

		r.Get("/{id}", applicationHandler.GetAppByID)
		r.Put("/{id}", applicationHandler.UpdateApp)
		r.Patch("/{id}/status", applicationHandler.UpdateStatus)
		r.Delete("/{id}", applicationHandler.DeleteApp)

		r.Get("/{id}/notes", notesHandler.GetNotes)
		r.Post("/{id}/notes", notesHandler.CreateNote)
		r.Post("/{id}/followups", followupsHandler.CreateFollowUp)
	})

	router.Route("/companies", func(r chi.Router) {
		r.Get("/", companiesHandler.GetAll)
		r.Get("/{id}", companiesHandler.GetByID)
		r.Post("/", companiesHandler.CreateComp)
		r.Put("/{id}", companiesHandler.UpdateComp)
		r.Delete("/{id}", companiesHandler.DeleteComp)
		r.Post("/{id}/contacts", contactsHandler.CreateContact)
		r.Get("/{id}/contacts", contactsHandler.GetContacts)
	})

	router.Route("/followups", func(r chi.Router) {
		r.Get("/due", followupsHandler.GetDueFollowUps)
		r.Put("/{id}/complete", followupsHandler.CompleteFollowUp)
		r.Delete("/{id}", followupsHandler.DeleteFollowUp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := ":" + port

	srv := &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		log.Printf("Server listening at: %s", address)

		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("shutdown signal recieved")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	log.Println("waiting for active requests to finish")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	log.Println("HTTP server stopped")

}

func (app *applicationServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
