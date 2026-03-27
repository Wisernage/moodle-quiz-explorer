package importer

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"timo/internal/db"
	"timo/internal/model"
)

func TestImportMinimalBackupDir(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "moodle", "minimal")
	gdb, err := db.OpenMemory()
	require.NoError(t, err)
	err = ImportBackupDir(gdb, root)
	require.NoError(t, err)

	var quizzes []model.Quiz
	require.NoError(t, gdb.Find(&quizzes).Error)
	require.Len(t, quizzes, 1)
	require.Equal(t, int64(10), quizzes[0].MoodleQuizID)

	var slots []model.QuizSlot
	require.NoError(t, gdb.Order("slot").Find(&slots).Error)
	require.Len(t, slots, 2)

	var qs []model.Question
	require.NoError(t, gdb.Find(&qs).Error)
	require.Len(t, qs, 2)
	count := map[string]int{}
	for _, q := range qs {
		count[q.Qtype]++
	}
	require.Equal(t, 1, count["multichoice"])
	require.Equal(t, 1, count["truefalse"])
}
