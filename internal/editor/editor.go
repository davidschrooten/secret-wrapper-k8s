package editor

import (
	"fmt"
	"os"
	"os/exec"
)

// SelectEditor determines which editor to use based on CLI flag and environment variables
// Priority order: 1) flagValue, 2) $EDITOR, 3) $VISUAL, 4) default to "vi"
func SelectEditor(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	return "vi"
}

// LaunchEditor launches the specified editor with the given file path
// The function waits for the editor to exit and returns any error
func LaunchEditor(editor string, args ...string) error {
	cmd := exec.Command(editor, args...)

	// Connect stdin, stdout, stderr to allow interactive editing
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	return nil
}
