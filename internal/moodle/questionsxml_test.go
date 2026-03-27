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
