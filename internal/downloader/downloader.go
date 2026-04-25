// Package downloader provides utilities for downloading, verifying, and
// extracting release artifacts. It handles the full lifecycle: HTTP download,
// SHA-256 checksum verification, tar.gz extraction, and binary installation.
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

var httpClient = &http.Client{Timeout: 5 * time.Minute}

// FetchJSON performs a GET request to url and decodes the JSON response into T.
func FetchJSON[T any](ctx context.Context, url string) (T, error) { //nolint:ireturn // generic function
	var zero T

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return zero, fmt.Errorf("create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("fetch %s: %w: %d", url, ErrHTTPStatus, resp.StatusCode)
	}

	var result T

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return zero, fmt.Errorf("parse %s: %w", url, err)
	}

	return result, nil
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

	f, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}

	defer func() {
		cErr := f.Close()
		if cErr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", dst, cErr)
		}
	}()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}

	return nil
}

// InstallBinary copies a binary to destDir and makes it executable.
func InstallBinary(src, destDir, name string) (err error) {
	err = os.MkdirAll(destDir, 0o750)
	if err != nil {
		return fmt.Errorf("create %s: %w", destDir, err)
	}

	dst := filepath.Join(destDir, name)

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

	err = os.Chmod(dst, 0o750)
	if err != nil {
		return fmt.Errorf("chmod %s: %w", dst, err)
	}

	return nil
}
