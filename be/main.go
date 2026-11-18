// Command be는 ci-cd-lab의 API 서버다(D-2, D-3).
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// 빌드 때 -ldflags "-X main.commit=... -X main.builtAt=..."로 채운다.
var (
	commit  = "dev"
	builtAt = "unknown"
)

func main() {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	addr := ":" + envOr("PORT", "8080")
	log.Printf("listening on %s", addr)
	srv := &http.Server{Addr: addr, Handler: newServer(db), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(srv.ListenAndServe())
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func newServer(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// D-3: DB 연결까지 1초 안에 확인되어야 준비된 것으로 본다.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "db unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"commit": commit, "builtAt": builtAt})
	})

	mux.HandleFunc("GET /api/notes", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(), "SELECT id, body FROM notes ORDER BY id")
		if err != nil {
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		notes := []note{}
		for rows.Next() {
			var n note
			if err := rows.Scan(&n.ID, &n.Body); err != nil {
				http.Error(w, "scan failed", http.StatusInternalServerError)
				return
			}
			notes = append(notes, n)
		}
		writeJSON(w, http.StatusOK, notes)
	})

	mux.HandleFunc("POST /api/notes", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil || in.Body == "" {
			http.Error(w, "body required", http.StatusBadRequest)
			return
		}
		var n note
		err := db.QueryRowContext(r.Context(),
			"INSERT INTO notes (body) VALUES ($1) RETURNING id, body", in.Body).Scan(&n.ID, &n.Body)
		if err != nil {
			http.Error(w, "insert failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, n)
	})

	return mux
}

type note struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
