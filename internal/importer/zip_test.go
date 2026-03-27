package importer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"timo/internal/db"
	"timo/internal/model"
)

func TestImportMBZZip(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "moodle", "minimal")
	zpath := filepath.Join(t.TempDir(), "course.mbz")
	f, err := os.Create(zpath)
	require.NoError(t, err)
	zw := zip.NewWriter(f)
	require.NoError(t, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	}))
	require.NoError(t, zw.Close())
	require.NoError(t, f.Close())

	gdb, err := db.OpenMemory()
	require.NoError(t, err)
	require.NoError(t, ImportMBZ(gdb, zpath))
}

func TestImportMBZTarGz(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "moodle", "minimal")
	tgzPath := filepath.Join(t.TempDir(), "course.mbz")

	f, err := os.Create(tgzPath)
	require.NoError(t, err)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	require.NoError(t, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == "." {
			return nil
		}

		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
		}
		if info.IsDir() {
			hdr.Name += "/"
			hdr.Typeflag = tar.TypeDir
			hdr.Mode = 0o755
			return tw.WriteHeader(hdr)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hdr.Size = int64(len(data))
		hdr.Typeflag = tar.TypeReg
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	}))

	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	require.NoError(t, f.Close())

	gdb, err := db.OpenMemory()
	require.NoError(t, err)
	require.NoError(t, ImportMBZ(gdb, tgzPath))

	var quizzes []model.Quiz
	require.NoError(t, gdb.Find(&quizzes).Error)
	require.Len(t, quizzes, 1, "tar.gz import should produce one quiz")
}

func TestDetectFormat(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "moodle", "minimal")
	tmp := t.TempDir()

	zipPath := filepath.Join(tmp, "test.zip.mbz")
	f, err := os.Create(zipPath)
	require.NoError(t, err)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("moodle_backup.xml")
	_, _ = w.Write([]byte("<x/>"))
	require.NoError(t, zw.Close())
	require.NoError(t, f.Close())

	fmt1, err := detectFormat(zipPath)
	require.NoError(t, err)
	require.Equal(t, mbzZip, fmt1)

	tgzPath := filepath.Join(tmp, "test.tgz.mbz")
	f2, err := os.Create(tgzPath)
	require.NoError(t, err)
	gz := gzip.NewWriter(f2)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: ".ARCHIVE_INDEX", Size: 0, Typeflag: tar.TypeReg})
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	require.NoError(t, f2.Close())

	fmt2, err := detectFormat(tgzPath)
	require.NoError(t, err)
	require.Equal(t, mbzTgz, fmt2)

	_, err = detectFormat(filepath.Join(src, "moodle_backup.xml"))
	require.Error(t, err, "XML file should not be detected as valid .mbz format")
}
