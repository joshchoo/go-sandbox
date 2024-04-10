package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joshchoo/go-sandbox/httpserver/database"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

const dbFile = "file:./httpserver/db.sqlite"

const maxCacheSizeBytes = 1_000_000 // 1 MB

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	db, err := database.InitSQLiteDB(ctx, dbFile)
	if err != nil {
		return err
	}

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

	h.Handle("POST /cache", http.MaxBytesHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, err := db.ExecContext(ctx, `INSERT INTO blob_cache (key, data) VALUES (?, ?)`, header.Filename, fileBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(map[string]any{
			"id": id,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}), maxCacheSizeBytes))

	s := http.Server{
		Addr:    "localhost:8000",
		Handler: h,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			slog.Error(err.Error())
		}
	}()

	<-ctx.Done()
	slog.InfoContext(ctx, "Exit signal received. Shutting down server.")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}
