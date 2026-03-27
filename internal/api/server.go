package api

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"timo/internal/model"
)

//go:embed templates/*.html
var templateFS embed.FS

// Server serves HTTP handlers for quizzes.
type Server struct {
	DB *gorm.DB
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /quizzes", s.handleQuizList)
	mux.HandleFunc("GET /quizzes/{id}", s.handleQuizGet)
	mux.HandleFunc("POST /quizzes/{id}/attempts", s.handleAttemptPost)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	var qs []model.Quiz
	if err := s.DB.Order("id asc").Find(&qs).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFS(templateFS, "templates/list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, qs)
}

func (s *Server) handleQuizList(w http.ResponseWriter, r *http.Request) {
	var qs []model.Quiz
	if err := s.DB.Order("id asc").Find(&qs).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("format") == "json" || strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(qs)
		return
	}
	tmpl, err := template.ParseFS(templateFS, "templates/list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, qs)
}

func (s *Server) handleQuizGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	var q model.Quiz
	if err := s.DB.Preload("Slots", func(db *gorm.DB) *gorm.DB {
		return db.Order("slot asc")
	}).Preload("Slots.Question").Preload("Sections").Preload("Feedbacks").First(&q, id).Error; err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if r.URL.Query().Get("format") == "json" || strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(q)
		return
	}
	s.renderQuizPage(w, &q)
}

func (s *Server) handleAttemptPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	var body struct {
		Answers map[string]any `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, _ := json.Marshal(body.Answers)
	pa := model.PracticeAttempt{QuizID: uint(id), AnswersJSON: string(b)}
	if err := s.DB.Create(&pa).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"id": pa.ID, "ok": true})
}

// quizPageData is passed to quiz.html (embedded Quiz + JSON for attempt preview).
type quizPageData struct {
	*model.Quiz
	SlotsWire template.JS
}

func (s *Server) renderQuizPage(w http.ResponseWriter, q *model.Quiz) {
	tmpl := template.New("quiz.html").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"isUnsupported": func(qtype string) bool {
			switch strings.ToLower(strings.TrimSpace(qtype)) {
			case "multichoice", "truefalse", "shortanswer", "essay", "match", "description":
				return false
			default:
				return true
			}
		},
	})
	tmpl, err := tmpl.ParseFS(templateFS, "templates/quiz.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type slotWire struct {
		Slot         int             `json:"slot"`
		Qtype        string          `json:"qtype"`
		QuestionText string          `json:"questionText"`
		Payload      json.RawMessage `json:"payload"`
	}
	var wire []slotWire
	for _, sl := range q.Slots {
		sw := slotWire{Slot: sl.Slot}
		if sl.Question != nil {
			sw.Qtype = sl.Question.Qtype
			sw.QuestionText = sl.Question.QuestionText
			p := strings.TrimSpace(sl.Question.Payload)
			if p == "" {
				sw.Payload = json.RawMessage(`{}`)
			} else if json.Valid([]byte(p)) {
				sw.Payload = json.RawMessage(p)
			} else {
				sw.Payload = json.RawMessage(`{"qtype":"invalid"}`)
			}
		}
		wire = append(wire, sw)
	}
	slotsJSON, err := json.Marshal(wire)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, quizPageData{Quiz: q, SlotsWire: template.JS(slotsJSON)})
}
