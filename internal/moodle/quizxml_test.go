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
