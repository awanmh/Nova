package security_test

import (
	"path/filepath"
	"testing"

	"github.com/awanmh/Nova/internal/security"
)

func TestRedactor(t *testing.T) {
	r := security.NewRedactor()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Redact OpenAI key in env format",
			input:    "OPENAI_API_KEY=sk-1234567890abcdef1234567890abcdef",
			expected: "OPENAI_API_KEY=[REDACTED]",
		},
		{
			name:     "Redact private key",
			input:    "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQ...\n-----END RSA PRIVATE KEY-----",
			expected: "[REDACTED PRIVATE KEY]",
		},
		{
			name:     "Normal text unredacted",
			input:    "func main() { fmt.Println(\"hello\") }",
			expected: "func main() { fmt.Println(\"hello\") }",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := r.Redact(tc.input)
			if got != tc.expected {
				t.Fatalf("expected '%s', got '%s'", tc.expected, got)
			}
		})
	}
}

func TestValidateWorkspacePath(t *testing.T) {
	root := filepath.Join("testdata", "workspace")

	validPath, err := security.ValidateWorkspacePath(root, "src/main.go")
	if err != nil {
		t.Fatalf("unexpected error for valid path: %v", err)
	}
	if validPath == "" {
		t.Fatalf("expected resolved path, got empty")
	}

	_, err = security.ValidateWorkspacePath(root, "../../etc/passwd")
	if err == nil {
		t.Fatalf("expected security boundary error for traversal path, got nil")
	}
}

func TestIsSensitiveFile(t *testing.T) {
	sensitive := []string{".env", ".env.local", "server.pem", "id_rsa.key", "credentials.json"}
	for _, f := range sensitive {
		if !security.IsSensitiveFile(f) {
			t.Fatalf("expected '%s' to be marked as sensitive file", f)
		}
	}

	normal := []string{"main.go", "README.md", "config.yaml"}
	for _, f := range normal {
		if security.IsSensitiveFile(f) {
			t.Fatalf("expected '%s' to be marked as normal file", f)
		}
	}
}
