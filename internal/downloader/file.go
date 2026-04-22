package downloader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile copies the file at src to dst.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}

	defer func() { _ = in.Close() }()

	out, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}

	defer func() {
		cErr := out.Close()
		if cErr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", dst, cErr)
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}

	return nil
}

// InstallBinary copies a binary to destDir and makes it executable.
func InstallBinary(src, destDir, name string) error {
	err := os.MkdirAll(destDir, 0o750)
	if err != nil {
		return fmt.Errorf("create %s: %w", destDir, err)
	}

	dst := filepath.Join(destDir, name)

	err = CopyFile(src, dst)
	if err != nil {
		return err
	}

	err = os.Chmod(dst, 0o750)
	if err != nil {
		return fmt.Errorf("chmod %s: %w", dst, err)
	}

	return nil
}
