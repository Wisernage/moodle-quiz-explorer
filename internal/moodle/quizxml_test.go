package moodle

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseQuizActivityMinimal(t *testing.T) {
	p := filepath.Join("..", "..", "testdata", "moodle", "minimal", "activities", "quiz_1", "quiz.xml")
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	q, err := ParseQuizActivityXML(bytes.NewReader(b))
	require.NoError(t, err)
	require.Equal(t, int64(10), q.ID)
	require.Equal(t, "Demo quiz", q.Name)
	require.Len(t, q.QuestionInstances.Instances, 2)
	require.Equal(t, int64(5), q.QuestionInstances.Instances[0].QuestionBankEntryID())
}

func TestParseQuizActivityLegacyQuestionIDs(t *testing.T) {
	const xmlDoc = `<?xml version="1.0" encoding="UTF-8"?>
<activity id="66" modulename="quiz">
  <quiz id="66">
    <name>Legacy</name>
    <intro></intro>
    <introformat>1</introformat>
    <preferredbehaviour>deferredfeedback</preferredbehaviour>
    <shuffleanswers>1</shuffleanswers>
    <sumgrades>3.00000</sumgrades>
    <grade>10.00000</grade>
    <question_instances>
      <question_instance id="263">
        <slot>1</slot>
        <page>1</page>
        <requireprevious>0</requireprevious>
        <questionid>285</questionid>
        <maxmark>1.0000000</maxmark>
      </question_instance>
    </question_instances>
    <sections/>
    <feedbacks/>
    <overrides/>
    <grades/>
    <attempts/>
  </quiz>
</activity>`
	q, err := ParseQuizActivityXML(bytes.NewReader([]byte(xmlDoc)))
	require.NoError(t, err)
	require.Len(t, q.QuestionInstances.Instances, 1)
	require.Equal(t, int64(285), q.QuestionInstances.Instances[0].LegacyMoodleQuestionID)
	require.Equal(t, int64(0), q.QuestionInstances.Instances[0].QuestionBankEntryID())
}
