package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func checkFzf() error {
	// Check if fzf is available
	_, err := exec.LookPath("fzf")
	if err != nil {
		return errors.New("fzf (https://github.com/junegunn/fzf) is required for search functionality")
	}
	return nil
}

func find(dir string) (string, error) {
	cmd := exec.Command("fzf")

	cmd.Dir = dir

	// Connect stdin and stderr for interactive use, but capture stdout
	// so that we can get the selected file for further processing
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// Capture the selection
	output, err := cmd.Output()

	if err != nil {
		// fzf returns exit code 1 when user cancels (Ctrl+C), which is normal
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return "", nil
			}
		} else {
			return "", fmt.Errorf("running fzf: %w", err)
		}
	}

	return strings.TrimSpace(string(output)), nil
}
