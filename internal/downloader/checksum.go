package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrChecksumMismatch is returned when a file's SHA-256 does not match the expected value.
var ErrChecksumMismatch = errors.New("checksum mismatch")

// ErrChecksumNotFound is returned when a checksum file does not contain an entry
// for the requested filename.
var ErrChecksumNotFound = errors.New("checksum not found")

// VerifyChecksum checks that the SHA-256 of file matches expected (hex string).
func VerifyChecksum(file, expected string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open %s: %w", file, err)
	}

	defer func() {
		cErr := f.Close()
		if cErr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", file, cErr)
		}
	}()

	h := sha256.New()

	_, err = io.Copy(h, f)
	if err != nil {
		return fmt.Errorf("read %s: %w", file, err)
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expected) {
		return fmt.Errorf("%w: got %s, want %s", ErrChecksumMismatch, got, expected)
	}

	return nil
}

// ExtractChecksum reads a checksum file and returns the hex checksum for the
// given filename.
func ExtractChecksum(shaFile, filename string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(shaFile))
	if err != nil {
		return "", fmt.Errorf("read %s: %w", shaFile, err)
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.TrimPrefix(parts[1], "*") == filename {
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrChecksumNotFound, filename)
}
