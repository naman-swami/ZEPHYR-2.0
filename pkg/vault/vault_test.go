package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretDetection(t *testing.T) {
	tests := []struct {
		key      string
		val      string
		isSecret bool
	}{
		{"PORT", "8080", false},
		{"HOST", "127.0.0.1", false},
		{"NODE_ENV", "production", false},
		{"OPENAI_API_KEY", "sk-proj-1234567890abcdef123456", true},
		{"GITHUB_TOKEN", "ghp_abcdefghijklmnopqrstuvwxyz123456", true},
		{"DATABASE_PASSWORD", "supersecret123", true},
		{"AWS_SECRET_ACCESS_KEY", "AKIAIOSFODNN7EXAMPLE", true},
	}

	for _, tt := range tests {
		got, _ := IsSecretVariable(tt.key, tt.val)
		if got != tt.isSecret {
			t.Errorf("IsSecretVariable(%q, %q) = %v, want %v", tt.key, tt.val, got, tt.isSecret)
		}
	}
}

func TestCloakEnvContent(t *testing.T) {
	raw := `# Application Config
PORT=3000
DEBUG=false

# Secrets
OPENAI_API_KEY=sk-proj-9876543210abcdef987654
DB_PASS=mySuperSecretPass!
`
	res, err := CloakEnvContent(".env", raw)
	if err != nil {
		t.Fatalf("CloakEnvContent failed: %v", err)
	}

	if res.SecretCount != 2 {
		t.Errorf("expected 2 secrets, got %d", res.SecretCount)
	}
	if res.PublicCount != 2 {
		t.Errorf("expected 2 public vars, got %d", res.PublicCount)
	}

	// Verify cloaked content does NOT contain raw secret
	if strings.Contains(res.CloakedContent, "sk-proj-9876543210abcdef987654") {
		t.Fatalf("cloaked content leaked OpenAI API key!")
	}
	if strings.Contains(res.CloakedContent, "mySuperSecretPass!") {
		t.Fatalf("cloaked content leaked DB password!")
	}

	// Verify placeholder exists
	if !strings.Contains(res.CloakedContent, "ZEPHYR_VAULT_SECRET_") {
		t.Fatalf("expected synthetic placeholder token in cloaked content")
	}

	// Verify public config is preserved
	if !strings.Contains(res.CloakedContent, "PORT=3000") {
		t.Errorf("expected PORT=3000 to be preserved")
	}
}

func TestBidirectionalMerge(t *testing.T) {
	existingEnv := `# Existing User Configuration
PORT=3000
OPENAI_API_KEY=sk-proj-9876543210abcdef987654
DB_PASS=mySuperSecretPass!
`
	// AI receives the cloaked representation and wants to add REDIS_URL and change PORT to 8080
	aiProposedEnv := `# Existing User Configuration
PORT=8080
OPENAI_API_KEY=ZEPHYR_VAULT_SECRET_12345678
DB_PASS=ZEPHYR_VAULT_SECRET_87654321
REDIS_URL=redis://localhost:6379
`
	res, err := MergeEnvWithAI(".env", existingEnv, aiProposedEnv)
	if err != nil {
		t.Fatalf("MergeEnvWithAI failed: %v", err)
	}

	if res.PreservedSecrets != 2 {
		t.Errorf("expected 2 preserved secrets, got %d", res.PreservedSecrets)
	}
	if len(res.AddedVariables) != 1 || res.AddedVariables[0] != "REDIS_URL" {
		t.Errorf("expected REDIS_URL added, got %v", res.AddedVariables)
	}

	// Verify authentic secrets were restored into the merged content
	if !strings.Contains(res.MergedContent, "OPENAI_API_KEY=sk-proj-9876543210abcdef987654") {
		t.Errorf("expected authentic OpenAI key restored in merged content")
	}
	if !strings.Contains(res.MergedContent, "DB_PASS=mySuperSecretPass!") {
		t.Errorf("expected authentic DB password restored in merged content")
	}
	if !strings.Contains(res.MergedContent, "PORT=8080") {
		t.Errorf("expected updated PORT=8080 in merged content")
	}
	if !strings.Contains(res.MergedContent, "REDIS_URL=redis://localhost:6379") {
		t.Errorf("expected added REDIS_URL in merged content")
	}
}

func TestAuditWorkspace(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zephyr_vault_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	envContent := "PORT=8080\nAPI_KEY=sk-proj-testkey1234567890123456\n"
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0600)

	report, err := AuditWorkspace(tempDir)
	if err != nil {
		t.Fatalf("AuditWorkspace failed: %v", err)
	}

	if report.TotalSecrets != 1 {
		t.Errorf("expected 1 exposed secret, got %d", report.TotalSecrets)
	}
	if report.ProtectedStatus != "VULNERABLE" {
		t.Errorf("expected VULNERABLE status, got %s", report.ProtectedStatus)
	}
}
