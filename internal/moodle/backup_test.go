package moodle

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMoodleBackupMinimal(t *testing.T) {
	p := filepath.Join("..", "..", "testdata", "moodle", "minimal", "moodle_backup.xml")
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	refs, meta, err := ParseMoodleBackupXML(bytes.NewReader(b))
	require.NoError(t, err)
	require.Len(t, refs, 1)
	require.Equal(t, "activities/quiz_1", refs[0].Directory)
	require.Equal(t, "2024042200", meta.MoodleVersion)
}

func TestParseMoodleBackupFlatActivities(t *testing.T) {
	// Moodle 2.9+ style: quizzes listed under information/contents/activities (not under sections).
	const xmlDoc = `<?xml version="1.0" encoding="UTF-8"?>
<moodle_backup>
  <information>
    <moodle_version>2020110903.07</moodle_version>
    <contents>
      <activities>
        <activity>
          <moduleid>1</moduleid>
          <modulename>page</modulename>
          <title>Ignore</title>
          <directory>activities/page_1</directory>
        </activity>
        <activity>
          <moduleid>537</moduleid>
          <modulename>quiz</modulename>
          <title>Quiz A</title>
          <directory>activities/quiz_537</directory>
        </activity>
      </activities>
    </contents>
  </information>
</moodle_backup>`
	refs, meta, err := ParseMoodleBackupXML(bytes.NewReader([]byte(xmlDoc)))
	require.NoError(t, err)
	require.Len(t, refs, 1)
	require.Equal(t, int64(537), refs[0].ModuleID)
	require.Equal(t, "activities/quiz_537", refs[0].Directory)
	require.Equal(t, "Quiz A", refs[0].Title)
	require.Equal(t, "2020110903.07", meta.MoodleVersion)
}
