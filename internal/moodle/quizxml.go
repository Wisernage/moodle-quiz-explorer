package moodle

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// QuizXML mirrors mod_quiz backup XML under the quiz element.
type QuizXML struct {
	ID   int64 `xml:"id,attr"`
	Name string `xml:"name"`
	Intro string `xml:"intro"`
	IntroFormat int `xml:"introformat"`

	TimeOpen    int64 `xml:"timeopen"`
	TimeClose   int64 `xml:"timeclose"`
	TimeLimit   int64 `xml:"timelimit"`
	AttemptsNum int   `xml:"attempts_number"`
	GradeMethod int   `xml:"grademethod"`
	PreferredBehaviour string `xml:"preferredbehaviour"`
	ShuffleAnswers     string `xml:"shuffleanswers"`
	SumGrades          float64 `xml:"sumgrades"`
	Grade     float64 `xml:"grade"`

	QuestionInstances struct {
		Instances []QuestionInstanceXML `xml:"question_instance"`
	} `xml:"question_instances"`

	Sections struct {
		Sections []QuizSectionXML `xml:"section"`
	} `xml:"sections"`

	Feedbacks struct {
		Feedbacks []QuizFeedbackXML `xml:"feedback"`
	} `xml:"feedbacks"`

	QuizGradeItems struct {
		Items []QuizGradeItemXML `xml:"quiz_grade_item"`
	} `xml:"quiz_grade_items"`

	Overrides struct {
		Overrides []QuizOverrideXML `xml:"override"`
	} `xml:"overrides"`

	Grades struct {
		Grades []QuizGradeRowXML `xml:"grade"`
	} `xml:"grades"`

	Attempts struct {
		Attempts []QuizAttemptXML `xml:"attempt"`
	} `xml:"attempts"`
}

// QuestionInstanceXML is one quiz slot (question_instance in backup).
type QuestionInstanceXML struct {
	ID              int64   `xml:"id,attr"`
	Slot            int     `xml:"slot"`
	Page            int     `xml:"page"`
	MaxMark         float64 `xml:"maxmark"`
	QuizGradeItemID *int64  `xml:"quizgradeitemid"`
	// Legacy Moodle 3.x / pre–question-bank-reference: points to questions.xml <question id="…">.
	LegacyMoodleQuestionID int64 `xml:"questionid"`

	References []QuestionReferenceXML `xml:"question_reference"`
}

// QuestionReferenceXML holds bank entry id (Moodle 4+).
type QuestionReferenceXML struct {
	ID                  int64 `xml:"id,attr"`
	QuestionBankEntryID int64 `xml:"questionbankentryid"`
	Version             int   `xml:"version"`
}

// QuizSectionXML mirrors quiz_sections.
type QuizSectionXML struct {
	ID              int64  `xml:"id,attr"`
	FirstSlot       int    `xml:"firstslot"`
	Heading         string `xml:"heading"`
	ShuffleQuestions string `xml:"shufflequestions"`
}

// QuizFeedbackXML mirrors quiz_feedback.
type QuizFeedbackXML struct {
	ID                int64   `xml:"id,attr"`
	FeedbackText      string  `xml:"feedbacktext"`
	FeedbackTextFormat int    `xml:"feedbacktextformat"`
	MinGrade          float64 `xml:"mingrade"`
	MaxGrade          float64 `xml:"maxgrade"`
}

// QuizGradeItemXML mirrors quiz_grade_items.
type QuizGradeItemXML struct {
	ID       int64  `xml:"id,attr"`
	SortOrder int   `xml:"sortorder"`
	Name     string `xml:"name"`
}

// QuizOverrideXML mirrors quiz_overrides.
type QuizOverrideXML struct {
	ID        int64  `xml:"id,attr"`
	UserID    *int64 `xml:"userid"`
	GroupID   *int64 `xml:"groupid"`
	TimeOpen  *int64 `xml:"timeopen"`
	TimeClose *int64 `xml:"timeclose"`
	TimeLimit *int64 `xml:"timelimit"`
	Attempts  *int   `xml:"attempts"`
	Password  string `xml:"password"`
}

// QuizGradeRowXML mirrors quiz_grades.
type QuizGradeRowXML struct {
	ID           int64   `xml:"id,attr"`
	UserID       int64   `xml:"userid"`
	GradeVal     float64 `xml:"gradeval"`
	TimeModified *int64  `xml:"timemodified"`
}

// QuizAttemptXML mirrors quiz_attempts (subset).
type QuizAttemptXML struct {
	ID              int64   `xml:"id,attr"`
	UserID          int64   `xml:"userid"`
	AttemptNum      int     `xml:"attemptnum"`
	State           string  `xml:"state"`
	TimeStart       *int64  `xml:"timestart"`
	TimeFinish      *int64  `xml:"timefinish"`
	SumGrades       float64 `xml:"sumgrades"`
}

// ParseQuizActivityXML parses quiz.xml (finds the first <quiz> element; works inside <activity> or as root).
func ParseQuizActivityXML(r io.Reader) (*QuizXML, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil, fmt.Errorf("quiz xml: no quiz element")
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "quiz" {
			continue
		}
		var q QuizXML
		if err := dec.DecodeElement(&q, &se); err != nil {
			return nil, fmt.Errorf("decode quiz: %w", err)
		}
		if q.ID == 0 && strings.TrimSpace(q.Name) == "" {
			return nil, fmt.Errorf("quiz xml: empty quiz")
		}
		return &q, nil
	}
}

// QuestionBankEntryID returns the first question bank entry id for a slot, or 0.
func (q *QuestionInstanceXML) QuestionBankEntryID() int64 {
	if len(q.References) == 0 {
		return 0
	}
	return q.References[0].QuestionBankEntryID
}

// Version returns first reference version or 1.
func (q *QuestionInstanceXML) Version() int {
	if len(q.References) == 0 {
		return 1
	}
	v := q.References[0].Version
	if v == 0 {
		return 1
	}
	return v
}

// ParseFloat parses Moodle float strings.
func ParseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
