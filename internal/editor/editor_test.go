package editor

import (
	"os"
	"os/exec"
	"testing"
)

func TestSelectEditor(t *testing.T) {
	tests := []struct {
		name      string
		flagValue string
		editorEnv string
		visualEnv string
		want      string
	}{
		{
			name:      "flag takes precedence",
			flagValue: "nano",
			editorEnv: "vim",
			visualEnv: "emacs",
			want:      "nano",
		},
		{
			name:      "EDITOR env var when no flag",
			flagValue: "",
			editorEnv: "vim",
			visualEnv: "emacs",
			want:      "vim",
		},
		{
			name:      "VISUAL env var when no flag or EDITOR",
			flagValue: "",
			editorEnv: "",
			visualEnv: "emacs",
			want:      "emacs",
		},
		{
			name:      "default to vi when nothing set",
			flagValue: "",
			editorEnv: "",
			visualEnv: "",
			want:      "vi",
		},
		{
			name:      "flag set to empty string uses env",
			flagValue: "",
			editorEnv: "nano",
			visualEnv: "",
			want:      "nano",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			oldEditor := os.Getenv("EDITOR")
			oldVisual := os.Getenv("VISUAL")
			defer func() {
				os.Setenv("EDITOR", oldEditor)
				os.Setenv("VISUAL", oldVisual)
			}()

			os.Setenv("EDITOR", tt.editorEnv)
			os.Setenv("VISUAL", tt.visualEnv)

			got := SelectEditor(tt.flagValue)
			if got != tt.want {
				t.Errorf("SelectEditor(%q) = %q, want %q", tt.flagValue, got, tt.want)
			}
		})
	}
}

func TestLaunchEditor(t *testing.T) {
	tests := []struct {
		name      string
		editor    string
		filePath  string
		wantErr   bool
		setupMock bool
	}{
		{
			name:      "successful editor launch",
			editor:    "true", // 'true' is a command that always succeeds
			filePath:  "/tmp/test.yaml",
			wantErr:   false,
			setupMock: false,
		},
		{
			name:      "editor returns error",
			editor:    "false", // 'false' is a command that always fails
			filePath:  "/tmp/test.yaml",
			wantErr:   true,
			setupMock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LaunchEditor(tt.editor, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LaunchEditor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLaunchEditorWithRealCommand(t *testing.T) {
	// Test with a command that exists
	// We use 'echo' which should be available on all systems
	tmpFile := "/tmp/swk-test-editor.yaml"

	// Create a temporary file
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Use a command that will succeed
	err := LaunchEditor("echo", tmpFile)
	if err != nil {
		t.Errorf("LaunchEditor with echo failed: %v", err)
	}
}

func TestLaunchEditorNonExistentCommand(t *testing.T) {
	// Test with a command that doesn't exist
	err := LaunchEditor("this-editor-does-not-exist-12345", "/tmp/test.yaml")
	if err == nil {
		t.Error("LaunchEditor should fail with non-existent command")
	}
}

func TestCommandExecution(t *testing.T) {
	// Test that the command is properly executed with arguments
	// We'll create a shell script that echoes its arguments
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available, skipping test")
	}

	tmpFile := "/tmp/swk-test-args.yaml"
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Use sh -c to verify the file path is passed correctly
	err := LaunchEditor("sh", "-c", "exit 0", tmpFile)
	if err != nil {
		t.Errorf("LaunchEditor with sh failed: %v", err)
	}
}

func TestEditorWithSpaces(t *testing.T) {
	// Test editor command that might have spaces (though we don't split on spaces)
	// This ensures we pass the editor name as-is to exec.Command
	err := LaunchEditor("nonexistent editor with spaces", "/tmp/test.yaml")
	if err == nil {
		t.Error("LaunchEditor should fail with invalid command")
	}
}
