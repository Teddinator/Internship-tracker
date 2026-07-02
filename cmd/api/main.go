package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", homeHandler)
	r.Get("/applications", getApplication)
	r.Post("/applications", createApplication)
	r.Get("/applications/{id}", getApplicationsByID)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Internship Tracker API")
}

func getApplication(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "List applications")
}

func createApplication(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Create applications")
}

func getApplicationsByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprintln(w, "Get application with id: ", id)
}
