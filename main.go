package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

const port = "8080"

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	apiCfg := &apiConfig{}
	mux := http.NewServeMux()

	// APP
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./assets"))))
	// ADMIN
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	// GENERAL
	mux.HandleFunc("GET /healthz", healthHandler)
	// API
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.validateChirpHandler)

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

// Handlers
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	responseBody := []byte("OK")
	_, err := w.Write(responseBody)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	currentHits := cfg.fileserverHits.Load()
	responseBody := fmt.Appendf(nil, fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, currentHits))
	_, err := w.Write(responseBody)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.fileserverHits.Store(0)
	currentHits := cfg.fileserverHits.Load()
	responseBody := fmt.Appendf(nil, "Hits: %d", currentHits)
	_, err := w.Write(responseBody)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

type parameters struct {
	Body string `json:"body"`
}
type errorResponse struct {
	Error string `json:"error"`
}

type validResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func (cfg *apiConfig) validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	// Validate body length (140 chars)
	bodyLength := len(params.Body)
	fmt.Printf("the body length is: %v\n", bodyLength)
	if bodyLength > 140 {
		respondWithError(w, 500, "Chirp is too long")
		return
	}

	// hide profane words
	words := strings.Fields(params.Body)

	sanitizedWords := make([]string, len(words))
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if containsWord(profaneWords, lowerWord) {
			sanitizedWords[i] = "****"
		} else {
			sanitizedWords[i] = word
		}
	}

	// Add back spaces between words
	paragraph := strings.Join(sanitizedWords, " ")

	// Prepare repsonse
	respBody := validResponse{
		CleanedBody: paragraph,
	}
	respondWithJSON(w, 200, respBody)

}

// Middlewares
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Utilities

func containsWord(slice []string, targetWord string) bool {
	for _, word := range slice {
		if word == targetWord {
			return true
		}
	}
	return false
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	respBody := errorResponse{
		Error: msg,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Write(dat)
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	respBody := payload
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Write(dat)
	w.WriteHeader(200)
}
