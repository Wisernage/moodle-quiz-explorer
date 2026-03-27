package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"timo/internal/db"
	"timo/internal/importer"
)

func TestAPIHealthAndQuizzes(t *testing.T) {
	gdb, err := db.OpenMemory()
	require.NoError(t, err)
	root := filepath.Join("..", "..", "testdata", "moodle", "minimal")
	require.NoError(t, importer.ImportBackupDir(gdb, root))

	srv := &Server{DB: gdb}
	mux := http.NewServeMux()
	srv.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "ok")

	req2 := httptest.NewRequest(http.MethodGet, "/quizzes?format=json", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)
	var list []struct {
		ID uint `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &list))
	require.Len(t, list, 1)

	id := list[0].ID
	req3 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/quizzes/%d?format=json", id), nil)
	rec3 := httptest.NewRecorder()
	mux.ServeHTTP(rec3, req3)
	require.Equal(t, http.StatusOK, rec3.Code)

	body := `{"answers":{"1":"a"}}`
	req4 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/quizzes/%d/attempts", id), strings.NewReader(body))
	req4.Header.Set("Content-Type", "application/json")
	rec4 := httptest.NewRecorder()
	mux.ServeHTTP(rec4, req4)
	require.Equal(t, http.StatusCreated, rec4.Code)
}
