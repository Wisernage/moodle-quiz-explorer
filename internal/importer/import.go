package importer

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"

	"timo/internal/model"
	"timo/internal/moodle"
)

type mbzFormat int

const (
	mbzZip  mbzFormat = iota
	mbzTgz            // Moodle tgz_packer: gzip-compressed tar with .ARCHIVE_INDEX
)

// detectFormat inspects the first bytes to determine whether the .mbz is ZIP or tar.gz.
func detectFormat(path string) (mbzFormat, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("open backup: %w", err)
	}
	if fi.IsDir() {
		return 0, fmt.Errorf("%q is a directory, not an .mbz file — use import-dir instead", path)
	}
	if fi.Size() < 10 {
		return 0, fmt.Errorf("%q is too small (%d bytes) to be a valid Moodle backup", path, fi.Size())
	}

	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	hdr := make([]byte, 4)
	if _, err := io.ReadFull(f, hdr); err != nil {
		return 0, fmt.Errorf("read %q: %w", path, err)
	}

	if bytes.HasPrefix(hdr, []byte("PK")) {
		return mbzZip, nil
	}

	if hdr[0] == 0x1f && hdr[1] == 0x8b {
		return mbzTgz, nil
	}

	return 0, fmt.Errorf("%q does not look like a Moodle .mbz (first bytes % x). Expected ZIP (PK) or gzip (1F 8B)", path, hdr)
}

func parseShuffle(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "1" || s == "true" || s == "yes"
}

// ImportMBZ extracts the .mbz (ZIP or tar.gz) to a temp dir and imports into db.
func ImportMBZ(db *gorm.DB, mbzPath string) error {
	format, err := detectFormat(mbzPath)
	if err != nil {
		return err
	}

	dir, err := os.MkdirTemp("", "timo-mbz-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	switch format {
	case mbzZip:
		if err := Unzip(mbzPath, dir); err != nil {
			return fmt.Errorf("unzip: %w", err)
		}
	case mbzTgz:
		if err := UntarGz(mbzPath, dir); err != nil {
			return fmt.Errorf("untar: %w", err)
		}
	}
	return ImportBackupDir(db, dir)
}

// clearAllQuizData removes every quiz, question, and related row so a full import
// reflects only the current backup (backups with no quiz activities yield an empty DB).
func clearAllQuizData(tx *gorm.DB) error {
	sess := tx.Session(&gorm.Session{AllowGlobalUpdate: true})
	// Order: children before parents (FK). Course last.
	for _, m := range []any{
		&model.PracticeAttempt{},
		&model.QuizSlot{},
		&model.QuizSection{},
		&model.QuizFeedback{},
		&model.QuizGradeItem{},
		&model.QuizGrade{},
		&model.QuizOverride{},
		&model.QuizOverviewRegrade{},
		&model.QuizAttempt{},
		&model.QuizStatistic{},
		&model.Quiz{},
		&model.Question{},
		&model.Course{},
	} {
		if err := sess.Unscoped().Delete(m).Error; err != nil {
			return fmt.Errorf("clear %T: %w", m, err)
		}
	}
	return nil
}

// ImportBackupDir imports from an extracted Moodle backup root directory.
func ImportBackupDir(db *gorm.DB, root string) error {
	refs, _, err := moodle.ParseMoodleBackupFile(filepath.Join(root, "moodle_backup.xml"))
	if err != nil {
		return fmt.Errorf("moodle_backup.xml: %w", err)
	}

	questionsPath := filepath.Join(root, "questions.xml")
	var byQID map[int64]moodle.QuestionXML
	var byBank map[int64]moodle.QuestionXML
	if _, err := os.Stat(questionsPath); err == nil {
		f, err := os.Open(questionsPath)
		if err != nil {
			return err
		}
		byQID, byBank, err = moodle.ParseQuestionsFileCombined(f)
		_ = f.Close()
		if err != nil {
			return fmt.Errorf("questions.xml: %w", err)
		}
	} else {
		byQID = map[int64]moodle.QuestionXML{}
		byBank = map[int64]moodle.QuestionXML{}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := clearAllQuizData(tx); err != nil {
			return err
		}
		for _, ref := range refs {
			qpath := filepath.Join(root, filepath.FromSlash(ref.Directory), "quiz.xml")
			data, err := os.ReadFile(qpath)
			if err != nil {
				return fmt.Errorf("read %s: %w", qpath, err)
			}
			qx, err := moodle.ParseQuizActivityXML(strings.NewReader(string(data)))
			if err != nil {
				return fmt.Errorf("parse %s: %w", qpath, err)
			}
			if err := importOneQuiz(tx, qx, byQID, byBank); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(refs) == 0 {
		log.Printf("import: this backup has no quiz activities; database was cleared to match (0 quizzes). Include quizzes when exporting from Moodle, or use a course backup that contains quiz modules.")
	}
	return nil
}

func importOneQuiz(tx *gorm.DB, qx *moodle.QuizXML, byQID map[int64]moodle.QuestionXML, byBank map[int64]moodle.QuestionXML) error {
	// Delete existing quiz with same Moodle id for idempotent re-import.
	var existing model.Quiz
	if err := tx.Where("moodle_quiz_id = ?", qx.ID).First(&existing).Error; err == nil {
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizSlot{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizSection{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizFeedback{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizGradeItem{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizGrade{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizOverride{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizOverviewRegrade{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizAttempt{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.QuizStatistic{})
		tx.Where("quiz_id = ?", existing.ID).Delete(&model.PracticeAttempt{})
		tx.Delete(&existing)
	}

	q := model.Quiz{
		MoodleQuizID:      qx.ID,
		Name:              qx.Name,
		Intro:             qx.Intro,
		TimeOpen:          nz64(qx.TimeOpen),
		TimeClose:         nz64(qx.TimeClose),
		TimeLimit:         qx.TimeLimit,
		Attempts:          qx.AttemptsNum,
		Grademethod:       qx.GradeMethod,
		QuestionBehaviour: qx.PreferredBehaviour,
		ShuffleAnswers:    parseShuffle(qx.ShuffleAnswers),
		Sumgrades:         qx.SumGrades,
		Grade:             qx.Grade,
	}
	if err := tx.Create(&q).Error; err != nil {
		return err
	}

	for _, sec := range qx.Sections.Sections {
		s := model.QuizSection{
			QuizID:          q.ID,
			MoodleSectionID: sec.ID,
			FirstSlot:       sec.FirstSlot,
			Heading:         sec.Heading,
			Shuffle:         parseShuffle(sec.ShuffleQuestions),
		}
		if err := tx.Create(&s).Error; err != nil {
			return err
		}
	}

	for _, fb := range qx.Feedbacks.Feedbacks {
		f := model.QuizFeedback{
			QuizID:   q.ID,
			MinGrade: fb.MinGrade,
			MaxGrade: fb.MaxGrade,
			Text:     fb.FeedbackText,
		}
		if err := tx.Create(&f).Error; err != nil {
			return err
		}
	}

	for _, gi := range qx.QuizGradeItems.Items {
		row := model.QuizGradeItem{
			QuizID:            q.ID,
			MoodleGradeItemID: gi.ID,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}

	for _, ov := range qx.Overrides.Overrides {
		o := model.QuizOverride{
			QuizID:       q.ID,
			MoodleUserID: ov.UserID,
			MoodleGroupID: ov.GroupID,
			TimeOpen:     ov.TimeOpen,
			TimeClose:    ov.TimeClose,
			Attempts:     ov.Attempts,
		}
		if err := tx.Create(&o).Error; err != nil {
			return err
		}
	}

	for _, g := range qx.Grades.Grades {
		row := model.QuizGrade{
			QuizID:       q.ID,
			MoodleUserID: g.UserID,
			Grade:        g.GradeVal,
			TimeModified: g.TimeModified,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}

	for _, at := range qx.Attempts.Attempts {
		row := model.QuizAttempt{
			QuizID:          q.ID,
			MoodleAttemptID: at.ID,
			MoodleUserID:    at.UserID,
			State:           at.State,
			TimeStart:       at.TimeStart,
			TimeFinish:      at.TimeFinish,
			SumGrades:       at.SumGrades,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}

	// Ensure Question rows for each unique bank entry referenced by this quiz.
	bankSeen := make(map[int64]struct{})
	for _, inst := range qx.QuestionInstances.Instances {
		bankID := inst.QuestionBankEntryID()
		if bankID == 0 {
			continue
		}
		bankSeen[bankID] = struct{}{}
	}
	questionByBank := make(map[int64]uint)
	for bankID := range bankSeen {
		id, err := ensureQuestion(tx, bankID, byQID, byBank)
		if err != nil {
			return err
		}
		questionByBank[bankID] = id
	}

	for _, inst := range qx.QuestionInstances.Instances {
		bankID := inst.QuestionBankEntryID()
		qid := questionByBank[bankID]
		slot := model.QuizSlot{
			QuizID:                q.ID,
			Slot:                  inst.Slot,
			Page:                  inst.Page,
			MaxMark:               inst.MaxMark,
			MoodleBankEntryID:     bankID,
			MoodleQuestionVersion: inst.Version(),
		}
		if qid != 0 {
			slot.QuestionID = &qid
		}
		if err := tx.Create(&slot).Error; err != nil {
			return err
		}
	}

	return nil
}

func nz64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func ensureQuestion(tx *gorm.DB, bankID int64, byQID map[int64]moodle.QuestionXML, byBank map[int64]moodle.QuestionXML) (uint, error) {
	if bankID == 0 {
		return 0, nil
	}
	var ex model.Question
	if err := tx.Where("moodle_bank_entry_id = ?", bankID).First(&ex).Error; err == nil {
		return ex.ID, nil
	}

	qxml, ok := byBank[bankID]
	if !ok {
		qxml, ok = byQID[bankID]
	}
	if !ok {
		mq := model.Question{
			MoodleBankEntryID: bankID,
			Qtype:             "missingtype",
			Name:              fmt.Sprintf("missing bank entry %d", bankID),
			Payload:           `{"qtype":"missingtype","missing":true}`,
		}
		if err := tx.Create(&mq).Error; err != nil {
			return 0, err
		}
		return mq.ID, nil
	}

	raw := qxml.RawFragment
	if raw == "" {
		raw = "<question/>"
	}
	payload := moodle.BuildPayloadJSON(qxml.Type, []byte(raw))
	mq := model.Question{
		MoodleBankEntryID: bankID,
		MoodleQuestionID:  qxml.ID,
		Qtype:             qxml.Type,
		Name:              qxml.Name,
		QuestionText:      qxml.QuestionText.Text,
		DefaultMark:       qxml.DefaultGrade,
		GeneralFeedback:   qxml.GeneralFeedback.Text,
		IDNumber:          qxml.IDNumber,
		Payload:           payload,
		RawXML:            raw,
	}
	if err := tx.Create(&mq).Error; err != nil {
		return 0, err
	}
	return mq.ID, nil
}

// Unzip extracts a zip archive to destDir.
func Unzip(src, destDir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(destDir, filepath.FromSlash(f.Name))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(fpath)
		if err != nil {
			_ = rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		_ = rc.Close()
		_ = out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// UntarGz extracts a gzip-compressed tar archive (.tar.gz / Moodle tgz_packer format) to destDir.
func UntarGz(src, destDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip open: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		name := hdr.Name
		if name == ".ARCHIVE_INDEX" {
			continue
		}

		target := filepath.Join(destDir, filepath.FromSlash(name))
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("tar entry %q escapes destination", name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				_ = out.Close()
				return err
			}
			_ = out.Close()
		}
	}
	return nil
}
