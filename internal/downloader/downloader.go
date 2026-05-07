// Package downloader handles HTTP download, checksum verification, and extraction.
package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ErrHTTPStatus is returned when an HTTP response has an unexpected status code.
var ErrHTTPStatus = errors.New("unexpected HTTP status")

const (
	httpTimeoutMinutes = 5
	dirPerm            = 0o750
	execPerm           = 0o750
)

var httpClient = &http.Client{Timeout: httpTimeoutMinutes * time.Minute}

// FetchJSON performs a GET request and decodes the JSON response into result.
func FetchJSON[T any](ctx context.Context, url string, result *T) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", url, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch %s: %w: %d", url, ErrHTTPStatus, resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return fmt.Errorf("parse %s: %w", url, err)
	}

	return nil
}

// Download fetches the resource at url and writes it to dst.
func Download(ctx context.Context, url, dst string) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %w: %d", url, ErrHTTPStatus, resp.StatusCode)
	}

	destFile, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}

	defer func() {
		cErr := destFile.Close()
		if cErr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", dst, cErr)
		}
	}()

	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}

	return nil
}

// InstallBinary copies a binary to destDir atomically and makes it executable.
func InstallBinary(src, destDir, name string) error {
	err := os.MkdirAll(destDir, dirPerm)
	if err != nil {
		return fmt.Errorf("create %s: %w", destDir, err)
	}

	dst := filepath.Join(destDir, name)

	srcFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}

	defer func() { _ = srcFile.Close() }()

	tmp, err := os.CreateTemp(destDir, ".devstrap-install-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tmpName := tmp.Name()

	defer func() { _ = os.Remove(tmpName) }()

	_, err = io.Copy(tmp, srcFile)
	if err != nil {
		_ = tmp.Close()

		return fmt.Errorf("write %s: %w", tmpName, err)
	}

	err = tmp.Chmod(execPerm)
	if err != nil {
		_ = tmp.Close()

		return fmt.Errorf("chmod %s: %w", tmpName, err)
	}

	err = tmp.Close()
	if err != nil {
		return fmt.Errorf("close %s: %w", tmpName, err)
	}

	err = os.Rename(tmpName, dst)
	if err != nil {
		return fmt.Errorf("rename to %s: %w", dst, err)
	}

	return nil
}
