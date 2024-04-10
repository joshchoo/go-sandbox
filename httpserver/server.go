package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	h := http.NewServeMux()

	h.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	h.HandleFunc("GET /assets/{name}", func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("name")
		path := filepath.Join("assets", filename)
		http.ServeFile(w, r, path)
	})

	s := http.Server{
		Addr:    "localhost:8000",
		Handler: h,
	}

	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
