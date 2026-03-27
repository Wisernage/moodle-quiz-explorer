package model

import (
	"time"

	"gorm.io/gorm"
)

// AllModels returns every model for AutoMigrate (order: parents before children).
func AllModels() []any {
	return []any{
		&Course{},
		&Quiz{},
		&QuizSection{},
		&QuizSlot{},
		&QuizFeedback{},
		&QuizGradeItem{},
		&QuizGrade{},
		&QuizOverride{},
		&QuizOverviewRegrade{},
		&QuizAttempt{},
		&QuizStatistic{},
		&PracticeAttempt{},
		&Question{},
	}
}

// PracticeAttempt stores a lightweight practice submission from the test UI (not Moodle-native).
type PracticeAttempt struct {
	ID     uint `gorm:"primaryKey"`
	QuizID uint `gorm:"index;not null"`

	AnswersJSON string `gorm:"type:text"`

	CreatedAt time.Time
}

// Course is an optional course stub from a backup.
type Course struct {
	ID              uint   `gorm:"primaryKey"`
	MoodleContextID int64  `gorm:"index"` // optional
	Name            string `gorm:"size:512"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	Quizzes []Quiz `gorm:"foreignKey:CourseID"`
}

// Quiz mirrors mod_quiz instance fields used in backups.
type Quiz struct {
	ID uint `json:"id" gorm:"primaryKey"`

	CourseID *uint `json:"courseId,omitempty" gorm:"index"`
	Course   *Course `json:"-"`

	// MoodleQuizID is the activity instance id from backup (quiz.id in XML).
	MoodleQuizID int64 `json:"moodleQuizId" gorm:"index;not null"`

	Name  string `json:"name" gorm:"size:1024"`
	Intro string `json:"intro" gorm:"type:text"`

	TimeOpen         *int64 `json:"timeOpen,omitempty"`
	TimeClose        *int64 `json:"timeClose,omitempty"`
	TimeLimit        int64  `json:"timeLimit"`
	Attempts         int    `json:"attempts"`
	Grademethod      int    `json:"gradeMethod"`
	QuestionBehaviour string `json:"questionBehaviour" gorm:"size:64"`
	ShuffleAnswers   bool   `json:"shuffleAnswers"`
	Sumgrades        float64 `json:"sumGrades"`
	Grade            float64 `json:"grade"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Sections        []QuizSection        `json:"sections,omitempty" gorm:"foreignKey:QuizID"`
	Slots           []QuizSlot           `json:"slots,omitempty" gorm:"foreignKey:QuizID"`
	Feedbacks       []QuizFeedback       `json:"feedbacks,omitempty" gorm:"foreignKey:QuizID"`
	GradeItems      []QuizGradeItem      `gorm:"foreignKey:QuizID"`
	Grades          []QuizGrade          `gorm:"foreignKey:QuizID"`
	Overrides       []QuizOverride       `gorm:"foreignKey:QuizID"`
	OverviewRegrades []QuizOverviewRegrade `gorm:"foreignKey:QuizID"`
	QuizAttempts     []QuizAttempt         `json:"quizAttempts,omitempty" gorm:"foreignKey:QuizID"`
	Statistics       []QuizStatistic       `json:"statistics,omitempty" gorm:"foreignKey:QuizID"`
}

// QuizSection mirrors quiz_sections.
type QuizSection struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	QuizID uint `json:"quizId" gorm:"index;not null"`
	Quiz   Quiz `json:"-"`

	MoodleSectionID int64 `json:"moodleSectionId" gorm:"index"`

	FirstSlot int    `json:"firstSlot"`
	Heading   string `json:"heading" gorm:"type:text"`
	Shuffle   bool   `json:"shuffle"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuizSlot mirrors quiz_slots — links to Question via MoodleQuestionID.
type QuizSlot struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	QuizID uint `json:"quizId" gorm:"index;not null"`
	Quiz   Quiz `json:"-"`

	Slot    int     `json:"slot"`
	Page    int     `json:"page"`
	MaxMark float64 `json:"maxMark"`

	// From question_reference in backup (Moodle 4+).
	MoodleBankEntryID     int64 `json:"moodleBankEntryId" gorm:"index"`
	MoodleQuestionVersion int   `json:"moodleQuestionVersion" gorm:"default:1"`

	QuestionID *uint     `json:"questionId,omitempty"`
	Question   *Question `json:"question,omitempty" gorm:"foreignKey:QuestionID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuizFeedback mirrors quiz_feedback (grade boundaries).
type QuizFeedback struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	QuizID uint `json:"quizId" gorm:"index;not null"`

	MinGrade float64 `json:"minGrade"`
	MaxGrade float64 `json:"maxGrade"`
	Text     string `json:"text" gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuizGradeItem mirrors quiz_grade_items.
type QuizGradeItem struct {
	ID     uint `gorm:"primaryKey"`
	QuizID uint `gorm:"index;not null"`

	MoodleGradeItemID int64
	NeedsUpdate       bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuizGrade mirrors quiz_grades (per-user final grade).
type QuizGrade struct {
	ID     uint `gorm:"primaryKey"`
	QuizID uint `gorm:"index;not null"`

	MoodleUserID int64 `gorm:"index"`
	Grade        float64
	TimeModified *int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuizOverride mirrors quiz_overrides.
type QuizOverride struct {
	ID     uint `gorm:"primaryKey"`
	QuizID uint `gorm:"index;not null"`

	MoodleUserID  *int64
	MoodleGroupID *int64

	TimeOpen  *int64
	TimeClose *int64
	Attempts  *int

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuizOverviewRegrade mirrors quiz_overview_regrades.
type QuizOverviewRegrade struct {
	ID     uint `gorm:"primaryKey"`
	QuizID uint `gorm:"index;not null"`

	MoodleQuestionID int64
	Regrade            bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuizAttempt mirrors quiz_attempts (simplified).
type QuizAttempt struct {
	ID     uint `gorm:"primaryKey"`
	QuizID uint `gorm:"index;not null"`

	MoodleAttemptID int64 `gorm:"index"`
	MoodleUserID    int64 `gorm:"index"`

	State       string `gorm:"size:32"`
	TimeStart   *int64
	TimeFinish  *int64
	SumGrades   float64
	// RawPayload stores extra attempt data as JSON for fidelity.
	RawPayload string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuizStatistic stores optional cached statistics from backup.
type QuizStatistic struct {
	ID     uint `gorm:"primaryKey"`
	QuizID uint `gorm:"index;not null"`

	Kind       string `gorm:"size:64;index"`
	RawPayload string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Question is a question bank entry (or inlined) with type-specific payload.
type Question struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// MoodleBankEntryID is set when resolved from question_bank_entry (Moodle 4+).
	MoodleBankEntryID int64 `json:"moodleBankEntryId" gorm:"index"`
	// MoodleQuestionID is the legacy `<question id="...">` when present.
	MoodleQuestionID int64 `json:"moodleQuestionId" gorm:"index"`

	Qtype              string `json:"qtype" gorm:"size:64;index;not null"`
	Name               string `json:"name" gorm:"size:1024"`
	QuestionText       string `json:"questionText" gorm:"type:text"`
	DefaultMark        float64 `json:"defaultMark"`
	GeneralFeedback    string `json:"generalFeedback" gorm:"type:text"`
	IDNumber           string `json:"idNumber" gorm:"size:256"`

	// Payload is JSON for typed fields; unsupported types store minimal fields here too.
	Payload string `json:"payload" gorm:"type:text"`
	// RawXML preserves the original question subtree for lossless import.
	RawXML string `json:"-" gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
