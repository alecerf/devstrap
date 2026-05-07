package downloader

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrUnsupportedArchive is returned when the archive format is not recognised.
var ErrUnsupportedArchive = errors.New("unsupported archive format")

// Extract extracts an archive into destDir, selecting the extraction method
// based on the file extension (.tar.gz/.tgz → tar, .zip → zip).
func Extract(ctx context.Context, archive, destDir string, stripComponents int) error {
	switch {
	case strings.HasSuffix(archive, ".tar.gz"), strings.HasSuffix(archive, ".tgz"):
		return ExtractTarGz(ctx, archive, destDir, stripComponents)
	case strings.HasSuffix(archive, ".zip"):
		return ExtractZip(archive, destDir, stripComponents)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedArchive, archive)
	}
}

// ExtractTarGz extracts a .tar.gz archive into destDir.
// If stripComponents > 0, that many leading path elements are removed.
func ExtractTarGz(ctx context.Context, archive, destDir string, stripComponents int) error {
	args := []string{"-xzf", archive, "-C", destDir}
	if stripComponents > 0 {
		args = append(args, fmt.Sprintf("--strip-components=%d", stripComponents))
	}

	cmd := exec.CommandContext(ctx, "tar", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("extract %s: %s: %w", archive, string(out), err)
	}

	return nil
}

// ExtractZip extracts a .zip archive into destDir.
// If stripComponents > 0, that many leading path elements are removed.
func ExtractZip(archive, destDir string, stripComponents int) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("open %s: %w", archive, err)
	}

	defer func() { _ = reader.Close() }()

	for _, entry := range reader.File {
		name := stripPath(entry.Name, stripComponents)
		if name == "" {
			continue
		}

		dest := filepath.Join(destDir, filepath.FromSlash(name))

		// Prevent path traversal.
		if !strings.HasPrefix(dest, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}

		if entry.FileInfo().IsDir() {
			err = os.MkdirAll(dest, dirPerm)
			if err != nil {
				return fmt.Errorf("mkdir %s: %w", dest, err)
			}

			continue
		}

		err = extractZipFile(entry, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(entry *zip.File, dest string) error {
	err := os.MkdirAll(filepath.Dir(dest), dirPerm)
	if err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
	}

	source, err := entry.Open()
	if err != nil {
		return fmt.Errorf("open entry %s: %w", entry.Name, err)
	}

	defer func() { _ = source.Close() }()

	out, err := os.OpenFile(filepath.Clean(dest), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, entry.Mode())
	if err != nil {
		return fmt.Errorf("create %s: %w", dest, err)
	}

	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, source) //nolint:gosec
	if err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}

	return nil
}

// stripPath removes the first n path components from a slash-separated path.
func stripPath(name string, n int) string {
	for range n {
		slash := strings.IndexByte(name, '/')
		if slash == -1 {
			return ""
		}

		name = name[slash+1:]
	}

	return name
}
