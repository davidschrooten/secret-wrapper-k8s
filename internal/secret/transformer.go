package secret

import (
	"encoding/base64"
	"fmt"

	"gopkg.in/yaml.v3"
)

// IsSecret checks if the given YAML is a Kubernetes Secret resource
func IsSecret(input []byte) bool {
	if len(input) == 0 {
		return false
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(input, &doc); err != nil {
		return false
	}

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return false
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return false
	}

	kind := findField(root, "kind")
	return kind != nil && kind.Value == "Secret"
}

// DecodeSecretData takes a Kubernetes Secret YAML and decodes all base64 values in the data section
func DecodeSecretData(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(input, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateSecret(&doc); err != nil {
		return nil, err
	}

	if err := transformData(&doc, decodeBase64); err != nil {
		return nil, err
	}

	output, err := marshalWithIndent(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return output, nil
}

// EncodeSecretData takes a Kubernetes Secret YAML with plaintext data and encodes values to base64
func EncodeSecretData(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(input, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateSecret(&doc); err != nil {
		return nil, err
	}

	if err := transformData(&doc, encodeBase64); err != nil {
		return nil, err
	}

	output, err := marshalWithIndent(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return output, nil
}

// validateSecret checks if the YAML is a valid Kubernetes Secret
func validateSecret(doc *yaml.Node) error {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return fmt.Errorf("invalid YAML document")
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("expected YAML mapping")
	}

	// Find and validate "kind" field
	kind := findField(root, "kind")
	if kind == nil || kind.Value != "Secret" {
		return fmt.Errorf("not a Secret resource")
	}

	return nil
}

// findField finds a field in a YAML mapping node
func findField(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}

	return nil
}

// transformData applies a transformation function to all values in the "data" section
func transformData(doc *yaml.Node, transform func(string) (string, error)) error {
	root := doc.Content[0]
	dataNode := findField(root, "data")

	if dataNode == nil || dataNode.Kind != yaml.MappingNode {
		// No data section or empty, nothing to transform
		return nil
	}

	// Transform each value in the data section
	for i := 1; i < len(dataNode.Content); i += 2 {
		valueNode := dataNode.Content[i]

		// Handle scalar values
		if valueNode.Kind == yaml.ScalarNode {
			transformed, err := transform(valueNode.Value)
			if err != nil {
				return fmt.Errorf("failed to transform key %q: %w", dataNode.Content[i-1].Value, err)
			}
			valueNode.Value = transformed
			// Preserve or set appropriate style for multiline strings
			if containsNewline(transformed) {
				valueNode.Style = yaml.LiteralStyle
			} else {
				valueNode.Style = 0 // default style
			}
		}
	}

	return nil
}

// decodeBase64 decodes a base64 string
func decodeBase64(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid base64: %w", err)
	}
	return string(decoded), nil
}

// encodeBase64 encodes a string to base64
func encodeBase64(plain string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(plain)), nil
}

// containsNewline checks if a string contains newline characters
func containsNewline(s string) bool {
	for _, c := range s {
		if c == '\n' {
			return true
		}
	}
	return false
}

// marshalWithIndent marshals YAML with 2-space indentation
func marshalWithIndent(node *yaml.Node) ([]byte, error) {
	var buf []byte
	encoder := yaml.NewEncoder(&bytesWriter{buf: &buf})
	encoder.SetIndent(2)
	if err := encoder.Encode(node); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

// bytesWriter implements io.Writer to accumulate bytes
type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
