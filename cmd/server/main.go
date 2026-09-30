package main

import (
	"io/fs"
	"log"
	"net"
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
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := net.JoinHostPort(host, port)
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

	log.Printf("Carcassonne listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
