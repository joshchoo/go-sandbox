package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	h := http.NewServeMux()

	h.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
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
