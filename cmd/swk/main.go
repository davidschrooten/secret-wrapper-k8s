package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davidschrooten/secret-wrapper-k8s/internal/editor"
	"github.com/davidschrooten/secret-wrapper-k8s/internal/secret"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run is the main entry point that can be tested
func run(args []string) error {
	editorFlag, filePath, err := parseArgs(args)
	if err != nil {
		return err
	}

	// Process the secret file (decode base64)
	tmpFile, cleanup, err := processSecretFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to process secret file: %w", err)
	}
	defer cleanup()

	// Select and launch editor
	editorCmd := editor.SelectEditor(editorFlag)
	if err := editor.LaunchEditor(editorCmd, tmpFile); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Finalize: encode the edited file and write back to original
	if err := finalizeSecretFile(filePath, tmpFile); err != nil {
		return fmt.Errorf("failed to finalize secret file: %w", err)
	}

	return nil
}

// parseArgs parses command-line arguments and returns editor flag and file path
func parseArgs(args []string) (string, string, error) {
	fs := flag.NewFlagSet("swk", flag.ContinueOnError)
	editorFlag := fs.String("editor", "", "Editor to use (overrides $EDITOR and $VISUAL)")
	fs.String("e", "", "Shorthand for -editor")

	if err := fs.Parse(args); err != nil {
		return "", "", err
	}

	// Check for -e flag
	if e := fs.Lookup("e"); e != nil && e.Value.String() != "" {
		*editorFlag = e.Value.String()
	}

	// Get positional argument (file path)
	if fs.NArg() == 0 {
		return "", "", fmt.Errorf("usage: swk [-editor EDITOR] FILE")
	}

	filePath := fs.Arg(0)
	return *editorFlag, filePath, nil
}

// processSecretFile reads the secret file, decodes base64 values, and writes to a temp file
// Returns the temp file path and a cleanup function
func processSecretFile(filePath string) (string, func(), error) {
	// Read original file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Decode base64 values
	decoded, err := secret.DecodeSecretData(data)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode secret: %w", err)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "swk-*.yaml")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Write decoded data to temp file
	if _, err := tmpFile.Write(decoded); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return "", nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Cleanup function to remove temp file
	cleanup := func() {
		os.Remove(tmpPath)
	}

	return tmpPath, cleanup, nil
}

// finalizeSecretFile reads the edited temp file, encodes values, and writes back to original
func finalizeSecretFile(originalPath, tmpPath string) error {
	// Read edited data
	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Encode base64 values
	encoded, err := secret.EncodeSecretData(edited)
	if err != nil {
		return fmt.Errorf("failed to encode secret: %w", err)
	}

	// Write back to original file
	if err := os.WriteFile(originalPath, encoded, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
