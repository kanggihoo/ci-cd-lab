package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// openTestDB는 실제 Postgres에 붙는다(D-7).
// CI에서 DB 주소가 비어 있으면 건너뛰지 않고 실패한다.
// 건너뜀이 성공처럼 보이는 함정(D-5)을 시험 안에서도 만들지 않기 위해서다.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL is empty in CI")
		}
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec("DROP TABLE IF EXISTS notes"); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob("migrations/*.up.sql")
	for _, f := range files {
		sqlText, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(sqlText)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	return db
}

func TestHealthz(t *testing.T) {
	srv := newServer(openTestDB(t))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestNotesAddAndList(t *testing.T) {
	srv := newServer(openTestDB(t))

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("POST", "/api/notes", strings.NewReader(`{"body":"hello"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/notes", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"body":"hello"`) {
		t.Fatalf("GET status = %d body = %s", rec.Code, rec.Body.String())
	}
}

func TestNotesRejectsEmptyBody(t *testing.T) {
	srv := newServer(openTestDB(t))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("POST", "/api/notes", strings.NewReader(`{"body":""}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestVersion(t *testing.T) {
	srv := newServer(nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/version", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"commit"`) {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}
