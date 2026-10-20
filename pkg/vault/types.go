package vault

import (
	"time"
)

// RiskLevel defines the severity level of a detected secret.
type RiskLevel string

const (
	RiskCritical RiskLevel = "CRITICAL" // High entropy credentials, private keys, live API tokens
	RiskHigh     RiskLevel = "HIGH"     // Database passwords, secret tokens
	RiskMedium   RiskLevel = "MEDIUM"   // Internal service tokens, auth headers
	RiskLow      RiskLevel = "LOW"      // Generic identifiers
)

// EnvEntry represents a parsed line from an environment file.
type EnvEntry struct {
	Key            string `json:"key"`
	Value          string `json:"value"`
	RawLine        string `json:"raw_line"`
	IsSecret       bool   `json:"is_secret"`
	IsCommentOrBlank bool `json:"is_comment_or_blank"`
	RedactedToken  string `json:"redacted_token,omitempty"`
	Risk           RiskLevel `json:"risk,omitempty"`
}

// CloakResult captures the outcome of redacting sensitive files for AI agents.
type CloakResult struct {
	OriginalPath   string            `json:"original_path"`
	CloakedContent string            `json:"cloaked_content"`
	TotalVariables int               `json:"total_variables"`
	SecretCount    int               `json:"secret_count"`
	PublicCount    int               `json:"public_count"`
	SecretMap      map[string]string `json:"-"` // Maps placeholder tokens to authentic secrets (kept in memory only)
}

// MergeResult captures the result of bidirectional shadow vault reconciliation.
type MergeResult struct {
	TargetFile         string    `json:"target_file"`
	Timestamp          time.Time `json:"timestamp"`
	PreservedSecrets   int       `json:"preserved_secrets"`
	AddedVariables     []string  `json:"added_variables"`
	UpdatedVariables   []string  `json:"updated_variables"`
	MergedContent      string    `json:"merged_content"`
}

// AuditFinding records an exposed secret detected in the workspace.
type AuditFinding struct {
	FilePath      string    `json:"file_path"`
	LineNumber    int       `json:"line_number"`
	Key           string    `json:"key"`
	Risk          RiskLevel `json:"risk"`
	MaskedPreview string    `json:"masked_preview"`
	Description   string    `json:"description"`
}

// VaultAuditReport summarizes workspace credential security.
type VaultAuditReport struct {
	WorkspaceRoot   string         `json:"workspace_root"`
	FilesAudited    int            `json:"files_audited"`
	TotalSecrets    int            `json:"total_secrets"`
	Findings        []AuditFinding `json:"findings"`
	ProtectedStatus string         `json:"protected_status"` // "SECURED", "VULNERABLE"
}
