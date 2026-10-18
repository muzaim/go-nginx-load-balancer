package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Response struct {
	Status     string `json:"status"`
	ServedBy   string `json:"served_by"`
	ServerPort string `json:"server_port"`
	Timestamp  string `json:"timestamp"`
	Data       []User `json:"data"`
}

func main() {
	serverName := getEnv("SERVER_NAME", "app-default")
	port := getEnv("PORT", "8001")
	logDir := getEnv("LOG_DIR", "./logs")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Gagal membuat direktori log: %v", err)
	}

	users := []User{
		{ID: 1, Name: "Budi Santoso", Email: "budi@example.com", Role: "Backend Engineer"},
		{ID: 2, Name: "Siti Aminah", Email: "siti@example.com", Role: "DevOps Engineer"},
		{ID: 3, Name: "Andi Pratama", Email: "andi@example.com", Role: "Frontend Developer"},
	}

	writeHitLog := func(r *http.Request) {
		now := time.Now().Format("2006-01-02 15:04:05")
		logEntry := fmt.Sprintf("[%s] HIT -> Server: %-6s (Port: %s) | Method: %-4s | Path: %-12s | Client: %s\n",
			now, serverName, port, r.Method, r.URL.Path, r.RemoteAddr)

		fmt.Print(logEntry)

		logFile := fmt.Sprintf("%s/hits.log", logDir)
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			_, _ = f.WriteString(logEntry)
		}
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		writeHitLog(r)

		resp := Response{
			Status:     "success",
			ServedBy:   serverName,
			ServerPort: port,
			Timestamp:  time.Now().Format(time.RFC3339),
			Data:       users,
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Served-By", fmt.Sprintf("%s:%s", serverName, port))
		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(resp)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/users", handler)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"status":"healthy","server":"%s","port":"%s"}`, serverName, port)
	})

	log.Printf("🚀 Server %s berjalan di port %s...", serverName, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server gagal berjalan: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
