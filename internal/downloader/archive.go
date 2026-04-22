package downloader

import (
	"context"
	"fmt"
	"os/exec"
)

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
