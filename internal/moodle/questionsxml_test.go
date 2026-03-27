package moodle

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseQuestionsMinimal(t *testing.T) {
	p := filepath.Join("..", "..", "testdata", "moodle", "minimal", "questions.xml")
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	byQ, byBank, err := ParseQuestionsFileCombined(bytes.NewReader(b))
	require.NoError(t, err)
	require.Contains(t, byQ, int64(5))
	require.Contains(t, byQ, int64(6))
	require.NotEmpty(t, byQ[5].RawFragment)
	_ = byBank
}

func TestEnrichLegacyMoodle3QuestionShape(t *testing.T) {
	const frag = `<question id="285">
		<name>Q1</name>
		<questiontext>&lt;p&gt;Stem here.&lt;/p&gt;</questiontext>
		<qtype>truefalse</qtype>
		<defaultmark>1.0000000</defaultmark>
	</question>`
	q := QuestionXML{ID: 285, RawFragment: frag}
	enrichLegacyQuestionXML(&q)
	require.Equal(t, "truefalse", q.Type)
	require.Equal(t, "Q1", q.Name)
	require.Contains(t, q.QuestionText.Text, "Stem here")
	require.InDelta(t, 1.0, q.DefaultGrade, 0.001)
}
