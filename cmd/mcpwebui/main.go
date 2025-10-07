package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort = 8080
	maxBodySize = 64 * 1024
)

type echoRequest struct {
	Message string `json:"message"`
}

type echoResponse struct {
	Result map[string]string `json:"result"`
}

func main() {
	log.SetFlags(0)

	assetsDir, err := resolveAssetsDir()
	if err != nil {
		log.Fatalf("failed to resolve assets directory: %v", err)
	}

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(assetsDir))
	mux.Handle("/", fileServer)
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/api/echo", handleEcho)

	port := resolvePort()
	addr := fmt.Sprintf(":%d", port)

	log.Printf("Serving MCP web UI from %s on 0.0.0.0:%d", assetsDir, port)
	log.Printf("Visit http://localhost:%d in your browser.", port)

	server := &http.Server{
		Addr:              addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server error: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBodySize))
	decoder.DisallowUnknownFields()

	var payload echoRequest
	if err := decoder.Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload.Message = strings.TrimSpace(payload.Message)
	if payload.Message == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("missing message"))
		return
	}

	response := echoResponse{
		Result: map[string]string{
			"tool":    "echo",
			"message": payload.Message,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func resolveAssetsDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("WEBUI_ASSETS_DIR")); override != "" {
		if stat, err := os.Stat(override); err == nil && stat.IsDir() {
			return override, nil
		}
		log.Printf("[warn] WEBUI_ASSETS_DIR=%s is not a directory; falling back to defaults", override)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	fallback := filepath.Join(cwd, "webui")
	if stat, err := os.Stat(fallback); err == nil && stat.IsDir() {
		return fallback, nil
	}

	return "", fmt.Errorf("unable to locate web UI assets directory")
}

func resolvePort() int {
	if value := strings.TrimSpace(os.Getenv("WEBUI_PORT")); value != "" {
		if port, err := strconv.Atoi(value); err == nil && port > 0 && port < 65536 {
			return port
		}
		log.Printf("[warn] Ignoring invalid WEBUI_PORT value: %s", value)
	}
	return defaultPort
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enableCORS(w)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to serialize JSON response: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(recorder, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, recorder.status, time.Since(start).Truncate(time.Millisecond))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
