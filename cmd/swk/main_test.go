package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantEditor   string
		wantFilePath string
		wantErr      bool
	}{
		{
			name:         "with editor flag and file",
			args:         []string{"-editor", "vim", "/tmp/secret.yaml"},
			wantEditor:   "vim",
			wantFilePath: "/tmp/secret.yaml",
			wantErr:      false,
		},
		{
			name:         "with -e flag and file",
			args:         []string{"-e", "nano", "/tmp/secret.yaml"},
			wantEditor:   "nano",
			wantFilePath: "/tmp/secret.yaml",
			wantErr:      false,
		},
		{
			name:         "no editor flag, just file",
			args:         []string{"/tmp/secret.yaml"},
			wantEditor:   "",
			wantFilePath: "/tmp/secret.yaml",
			wantErr:      false,
		},
		{
			name:         "no arguments",
			args:         []string{},
			wantEditor:   "",
			wantFilePath: "",
			wantErr:      true,
		},
		{
			name:         "only flags no file",
			args:         []string{"-editor", "vim"},
			wantEditor:   "",
			wantFilePath: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEditor, gotFile, err := parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if gotEditor != tt.wantEditor {
					t.Errorf("parseArgs() editor = %v, want %v", gotEditor, tt.wantEditor)
				}
				if gotFile != tt.wantFilePath {
					t.Errorf("parseArgs() file = %v, want %v", gotFile, tt.wantFilePath)
				}
			}
		})
	}
}

func TestProcessSecretFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	validSecret := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
type: Opaque
data:
  password: cGFzc3dvcmQxMjM=
`

	invalidYAML := `this is not valid yaml: [[[`

	notASecret := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`

	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
	}{
		{
			name:        "valid secret file",
			fileContent: validSecret,
			wantErr:     false,
		},
		{
			name:        "invalid YAML",
			fileContent: invalidYAML,
			wantErr:     true,
		},
		{
			name:        "not a secret",
			fileContent: notASecret,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, "test.yaml")
			if err := os.WriteFile(testFile, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			defer os.Remove(testFile)

			// Process the file
			tmpFile, cleanup, err := processSecretFile(testFile)
			if cleanup != nil {
				defer cleanup()
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("processSecretFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify temp file was created
				if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
					t.Errorf("processSecretFile() did not create temp file")
				}

				// Verify the temp file contains decoded data
				content, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("Failed to read temp file: %v", err)
				}

				// Should contain decoded password
				if !contains(content, []byte("password123")) {
					t.Error("processSecretFile() did not decode base64 values")
				}
			}
		})
	}
}

func TestProcessSecretFileNonExistent(t *testing.T) {
	_, _, err := processSecretFile("/nonexistent/file.yaml")
	if err == nil {
		t.Error("processSecretFile() should fail with non-existent file")
	}
}

func TestFinalizeSecretFile(t *testing.T) {
	tmpDir := t.TempDir()

	editedSecret := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
type: Opaque
data:
  password: newpassword456
  username: admin
`

	tests := []struct {
		name       string
		editedData string
		wantErr    bool
	}{
		{
			name:       "valid edited secret",
			editedData: editedSecret,
			wantErr:    false,
		},
		{
			name:       "invalid YAML after edit",
			editedData: "invalid: yaml: [[[",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original and temp files
			originalFile := filepath.Join(tmpDir, "original.yaml")
			tmpFile := filepath.Join(tmpDir, "temp.yaml")

			originalData := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
type: Opaque
data:
  password: cGFzc3dvcmQxMjM=
`

			if err := os.WriteFile(originalFile, []byte(originalData), 0644); err != nil {
				t.Fatalf("Failed to create original file: %v", err)
			}
			defer os.Remove(originalFile)

			if err := os.WriteFile(tmpFile, []byte(tt.editedData), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile)

			err := finalizeSecretFile(originalFile, tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("finalizeSecretFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify original file was updated with encoded data
				content, err := os.ReadFile(originalFile)
				if err != nil {
					t.Fatalf("Failed to read original file: %v", err)
				}

				// Should contain base64 encoded values
				if contains(content, []byte("newpassword456")) {
					t.Error("finalizeSecretFile() did not encode values")
				}
			}
		})
	}
}

func TestRun(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		fileContent string
		setupEnv    func()
		wantErr     bool
	}{
		{
			name:        "missing file argument",
			args:        []string{},
			fileContent: "",
			wantErr:     true,
		},
		{
			name:        "non-existent file",
			args:        []string{"/nonexistent/file.yaml"},
			fileContent: "",
			wantErr:     true,
		},
		{
			name:        "invalid secret file",
			args:        []string{filepath.Join(tmpDir, "invalid.yaml")},
			fileContent: "invalid: yaml: [[[",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileContent != "" {
				testFile := tt.args[0]
				if err := os.WriteFile(testFile, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(testFile)
			}

			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			err := run(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessSecretFileWriteError(t *testing.T) {
	// Test error when writing temp file fails
	// This is hard to trigger naturally, but we test what we can
	tmpDir := t.TempDir()
	validSecret := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
type: Opaque
data:
  password: cGFzc3dvcmQxMjM=
`
	testFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte(validSecret), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This should succeed normally
	tmpFile, cleanup, err := processSecretFile(testFile)
	if err != nil {
		t.Errorf("processSecretFile() should succeed: %v", err)
	}
	if cleanup != nil {
		cleanup()
	}
	// Verify temp file is cleaned up
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("Cleanup should remove temp file")
	}
}

func TestFinalizeSecretFileReadError(t *testing.T) {
	// Test error when reading edited file fails
	err := finalizeSecretFile("/tmp/original.yaml", "/nonexistent/temp.yaml")
	if err == nil {
		t.Error("finalizeSecretFile() should fail with non-existent temp file")
	}
}

func TestParseArgsEditorShorthand(t *testing.T) {
	// Ensure -e flag properly sets editor
	editor, file, err := parseArgs([]string{"-e", "emacs", "/tmp/test.yaml"})
	if err != nil {
		t.Fatalf("parseArgs() failed: %v", err)
	}
	if editor != "emacs" {
		t.Errorf("parseArgs() editor = %q, want %q", editor, "emacs")
	}
	if file != "/tmp/test.yaml" {
		t.Errorf("parseArgs() file = %q, want %q", file, "/tmp/test.yaml")
	}
}

// Helper function
func contains(data []byte, substr []byte) bool {
	for i := 0; i <= len(data)-len(substr); i++ {
		if string(data[i:i+len(substr)]) == string(substr) {
			return true
		}
	}
	return false
}
