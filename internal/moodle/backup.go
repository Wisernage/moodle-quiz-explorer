package moodle

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// MoodleBackupMeta is minimal data read from moodle_backup.xml.
type MoodleBackupMeta struct {
	MoodleVersion string
}

// QuizActivityRef points to a quiz activity directory inside the backup zip.
type QuizActivityRef struct {
	ModuleID  int64
	Directory string // e.g. activities/quiz_3
	Title     string
}

// backupContents matches either legacy layout (sections with nested activities) or
// Moodle 2.9+ course backups where all activities are listed under information/contents/activities.
type backupContents struct {
	Sections struct {
		Section []backupSection `xml:"section"`
	} `xml:"sections"`
	Activities backupActivities `xml:"activities"`
}

type backupRoot struct {
	XMLName     xml.Name `xml:"moodle_backup"`
	Information struct {
		MoodleVersion string         `xml:"moodle_version"`
		Contents      backupContents `xml:"contents"`
	} `xml:"information"`
	Contents backupContents `xml:"contents"`
}

type backupSection struct {
	SectionID  string           `xml:"sectionid"`
	Activities backupActivities `xml:"activities"`
}

type backupActivities struct {
	Activity []backupActivity `xml:"activity"`
}

type backupActivity struct {
	ModuleID   int64  `xml:"moduleid"`
	ModuleName string `xml:"modulename"`
	Directory  string `xml:"directory"`
	Title      string `xml:"title"`
}

// ParseMoodleBackupXML extracts quiz activity directory paths from moodle_backup.xml.
func ParseMoodleBackupXML(r io.Reader) ([]QuizActivityRef, *MoodleBackupMeta, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	var root backupRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, nil, fmt.Errorf("unmarshal moodle_backup: %w", err)
	}
	meta := &MoodleBackupMeta{MoodleVersion: strings.TrimSpace(root.Information.MoodleVersion)}
	var refs []QuizActivityRef
	appendQuiz := func(acts []backupActivity) {
		for _, act := range acts {
			if strings.TrimSpace(strings.ToLower(act.ModuleName)) != "quiz" {
				continue
			}
			dir := strings.TrimSpace(act.Directory)
			if dir == "" {
				continue
			}
			refs = append(refs, QuizActivityRef{
				ModuleID:  act.ModuleID,
				Directory: dir,
				Title:     strings.TrimSpace(act.Title),
			})
		}
	}
	// Moodle 2.9+ / typical exports: flat activity list under information/contents/activities.
	appendQuiz(root.Information.Contents.Activities.Activity)
	// Legacy / test fixtures: contents/sections/section/activities at moodle_backup root.
	for _, sec := range root.Contents.Sections.Section {
		appendQuiz(sec.Activities.Activity)
	}
	return refs, meta, nil
}

// ParseMoodleBackupFile opens a path and parses it.
func ParseMoodleBackupFile(path string) ([]QuizActivityRef, *MoodleBackupMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return ParseMoodleBackupXML(f)
}
