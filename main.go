package main

import (
	"fmt"
	"net/http"
)

const port = "8080"

func main() {
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("/healthz", healthHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("Server starting on port %s...\n", port)
	// Start the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Server failed to start: %v\n", err)
	} else if err == http.ErrServerClosed {
		fmt.Println("Server closed gracefully.")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	responseBody := []byte("OK")
	_, err := w.Write(responseBody)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}
