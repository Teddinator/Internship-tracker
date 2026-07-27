package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/teddinator/Internship-tracker/internal/applications"
	"github.com/teddinator/Internship-tracker/internal/companies"
	"github.com/teddinator/Internship-tracker/internal/notes"
)

type applicationServer struct {
	db *sql.DB
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	app := &applicationServer{
		db: db,
	}

	applicationHandler := applications.NewHandler(db)
	companiesHandler := companies.NewHandler(db)
	notesHandler := notes.NewHandler(db)

	router := chi.NewRouter()

	router.Route("/health", func(r chi.Router) {
		r.Get("/", app.healthHandler)
	})

	router.Route("/applications", func(r chi.Router) {
		r.Get("/", applicationHandler.GetAll)
		r.Get("/{id}", applicationHandler.GetAppByID)
		r.Post("/", applicationHandler.CreateApp)
		r.Put("/{id}", applicationHandler.UpdateApp)
		r.Delete("/{id}", applicationHandler.DeleteApp)
		r.Get("/{id}/notes", notesHandler.GetNotes)
		r.Post("/{id}/notes", notesHandler.CreateNote)
	})

	router.Route("/companies", func(r chi.Router) {
		r.Get("/", companiesHandler.GetAll)
		r.Get("/{id}", companiesHandler.GetByID)
		r.Post("/", companiesHandler.CreateComp)
		r.Put("/{id}", companiesHandler.UpdateComp)
		r.Delete("/{id}", companiesHandler.DeleteComp)
	})

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func (app *applicationServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
