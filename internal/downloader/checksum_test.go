package downloader_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecerf/devstrap/internal/downloader"
)

func writeTempFile(t *testing.T, data []byte) string {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "checksum-test-*")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmpFile.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	err = tmpFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	return tmpFile.Name()
}

func writeTempString(t *testing.T, content string) string {
	t.Helper()

	return writeTempFile(t, []byte(content))
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}

func TestVerifyChecksum(t *testing.T) {
	t.Parallel()

	content := []byte("hello world\n")
	validHex := sha256Hex(content)

	tests := []struct {
		name     string
		content  []byte
		expected string
		wantErr  error
		missing  bool
	}{
		{
			name:     "valid checksum",
			content:  content,
			expected: validHex,
		},
		{
			name:     "valid checksum uppercase",
			content:  content,
			expected: strings.ToUpper(validHex),
		},
		{
			name:     "mismatch",
			content:  content,
			expected: strings.Repeat("ab", sha256.Size),
			wantErr:  downloader.ErrChecksumMismatch,
		},
		{
			name:    "missing file",
			missing: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			filePath := buildVerifyPath(t, testCase.missing, testCase.content)
			err := downloader.VerifyChecksum(filePath, testCase.expected)
			assertVerifyResult(t, err, testCase.wantErr, testCase.missing)
		})
	}
}

func buildVerifyPath(t *testing.T, missing bool, content []byte) string {
	t.Helper()

	if missing {
		return filepath.Join(t.TempDir(), "nonexistent")
	}

	return writeTempFile(t, content)
}

func assertVerifyResult(t *testing.T, err, wantErr error, missing bool) {
	t.Helper()

	switch {
	case missing:
		if err == nil {
			t.Fatal("expected an error for missing file, got nil")
		}

		if errors.Is(err, downloader.ErrChecksumMismatch) {
			t.Fatal("missing file should not return ErrChecksumMismatch")
		}
	case wantErr != nil:
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	default:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestExtractChecksum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		filename string
		wantHash string
		wantErr  error
	}{
		{
			name:     "standard format",
			content:  "abc123  myfile.tar.gz\ndef456  other.tar.gz\n",
			filename: "myfile.tar.gz",
			wantHash: "abc123",
		},
		{
			name:     "star prefix BSD style",
			content:  "abc123 *myfile.tar.gz\n",
			filename: "myfile.tar.gz",
			wantHash: "abc123",
		},
		{
			name:     "not found",
			content:  "abc123  myfile.tar.gz\ndef456  other.tar.gz\n",
			filename: "missing.tar.gz",
			wantErr:  downloader.ErrChecksumNotFound,
		},
		{
			name:     "multiple entries pick correct",
			content:  "aaa111  first.tar.gz\nbbb222  second.tar.gz\nccc333  third.tar.gz\n",
			filename: "second.tar.gz",
			wantHash: "bbb222",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			shaPath := writeTempString(t, testCase.content)
			got, err := downloader.ExtractChecksum(shaPath, testCase.filename)
			assertExtractResult(t, got, err, testCase.wantHash, testCase.wantErr)
		})
	}
}

func assertExtractResult(t *testing.T, got string, err error, wantHash string, wantErr error) {
	t.Helper()

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}

		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != wantHash {
		t.Fatalf("got %q, want %q", got, wantHash)
	}
}
