# secret-wrapper-k8s (swk)

[![Go Version](https://img.shields.io/badge/Go-1.25.5-blue.svg)](https://golang.org/doc/go1.25)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidschrooten/secret-wrapper-k8s)](https://goreportcard.com/report/github.com/davidschrooten/secret-wrapper-k8s)
[![Tests](https://github.com/davidschrooten/secret-wrapper-k8s/actions/workflows/test.yml/badge.svg)](https://github.com/davidschrooten/secret-wrapper-k8s/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/badge/coverage-84.1%25-brightgreen.svg)](https://github.com/davidschrooten/secret-wrapper-k8s)

A CLI tool that makes editing Kubernetes Secrets easier by automatically decoding and encoding base64 values.

## The Problem

When editing Kubernetes Secrets with `kubectl edit secret`, all values in the `data` section are base64-encoded, making them difficult to read and edit. You typically have to:

1. Decode base64 values manually
2. Edit the plaintext
3. Re-encode to base64
4. Paste back into the YAML

This is tedious and error-prone.

## The Solution

`swk` (secret-wrapper-k8s) acts as an editor wrapper that:

1. Receives the Secret YAML from kubectl
2. Automatically decodes all base64 values in the `data` section
3. Opens your preferred editor with the decoded values
4. After you save and exit, re-encodes the values to base64
5. Returns the encoded YAML to kubectl

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/davidschrooten/secret-wrapper-k8s.git
cd secret-wrapper-k8s

# Build and install
make install

# Or specify a custom installation path
make install INSTALL_PATH=$HOME/.local/bin
```

### Manual Build

```bash
go build -o swk ./cmd/swk
```

## Usage

### With kubectl edit

Set `swk` as your editor when using `kubectl edit`:

```bash
# Use swk with your preferred editor (vim in this example)
EDITOR="swk --editor vim" kubectl edit secret my-secret

# Or using the short flag
EDITOR="swk -e nano" kubectl edit secret my-secret
```

### Editor Selection

`swk` determines which editor to use with the following priority:

1. `--editor` or `-e` flag (highest priority)
2. `$EDITOR` environment variable
3. `$VISUAL` environment variable
4. `vi` (default fallback)

### Examples

```bash
# Use vim
EDITOR="swk -e vim" kubectl edit secret database-credentials

# Use nano
EDITOR="swk -e nano" kubectl edit secret api-keys

# Use emacs
EDITOR="swk --editor emacs" kubectl edit secret tls-cert

# Use your default $EDITOR
EDITOR=swk kubectl edit secret my-secret
```

### Setting a Default

You can set `swk` as your default Kubernetes editor:

```bash
# In your ~/.bashrc or ~/.zshrc
export KUBE_EDITOR="swk -e vim"

# Now you can just run:
kubectl edit secret my-secret
```

## How It Works

1. kubectl calls `swk` with a temporary YAML file path
2. `swk` reads the file and validates it's a Kubernetes Secret
3. All values in the `data` section are decoded from base64 to plaintext
4. The decoded YAML is written to a new temporary file
5. Your chosen editor opens the temporary file
6. You edit the plaintext values and save
7. `swk` reads the edited file and encodes all `data` values back to base64
8. The encoded YAML is written back to kubectl's original temp file
9. kubectl applies the changes

**Note:** `swk` only processes the `data` section. The `stringData` section (if present) is left as-is, since it's already plaintext.

## Development

### Requirements

- Go 1.25.5 or later
- make (optional, but recommended)

### Building

```bash
make build
```

### Testing

```bash
# Run all tests
make test

# Generate coverage report
make coverage
```

### Linting

```bash
# Run linter (requires golangci-lint)
make lint

# Or just run go vet
make vet
```

### Project Structure

```
.
├── cmd/swk/              # Main application entry point
│   ├── main.go          # CLI orchestration
│   └── main_test.go     # Integration tests
├── internal/
│   ├── editor/          # Editor selection and launching
│   │   ├── editor.go
│   │   └── editor_test.go
│   └── secret/          # YAML transformation (base64 encode/decode)
│       ├── transformer.go
│       └── transformer_test.go
├── Makefile             # Build automation
└── README.md            # This file
```

## Examples

### Before (manual process)

```bash
$ kubectl get secret my-secret -o yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
type: Opaque
data:
  password: cGFzc3dvcmQxMjM=    # What does this mean?
  username: YWRtaW4=            # Have to decode manually
```

### After (with swk)

```bash
$ EDITOR="swk -e vim" kubectl edit secret my-secret

# Your editor opens with:
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
type: Opaque
data:
  password: password123          # Easy to read!
  username: admin                # Easy to edit!

# Edit the values directly, save, and exit
# swk automatically encodes them back to base64
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

David Schrooten