package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tudor/carcasstwo/internal/handlers"
	"github.com/tudor/carcasstwo/internal/hub"
	"github.com/tudor/carcasstwo/internal/lobby"
	"github.com/tudor/carcasstwo/web"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	l := lobby.NewManager()
	h := hub.New()
	s := handlers.New(l, h)

	static, err := fs.Sub(web.Static, "static")
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	s.Routes(r)

	log.Printf("Carcassonne listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
