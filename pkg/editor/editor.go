package editor

import (
	"fmt"
	"os"
	"os/exec"
)

// EditDraft opens a temporary file containing initialContent with the system's preferred editor
// and returns the modified content when the editor closes.
func EditDraft(initialContent string) (string, error) {
	tempFile, err := os.CreateTemp("", "jobber_draft_*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for editing: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.WriteString(initialContent); err != nil {
		return "", fmt.Errorf("failed to write initial content to temp file: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync temp file: %w", err)
	}

	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		for _, e := range []string{"nano", "vim", "vi"} {
			if _, err := exec.LookPath(e); err == nil {
				editorCmd = e
				break
			}
		}
	}
	if editorCmd == "" {
		return "", fmt.Errorf("no editor command (nano, vim, vi) found or configured via EDITOR environment variable")
	}

	cmd := exec.Command(editorCmd, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor execution failed: %w", err)
	}

	updatedBytes, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read updated draft from temp file: %w", err)
	}

	return string(updatedBytes), nil
}
