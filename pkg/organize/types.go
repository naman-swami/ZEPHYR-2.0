package organize

import (
	"time"
)

// FileCategory defines the high level categorization of a workspace file.
type FileCategory string

const (
	CategoryRootAnchor FileCategory = "root_anchor" // Invariant files (go.mod, tasks.json, README.md, LICENSE)
	CategoryDocs       FileCategory = "docs"        // Markdown, text documentation, specifications
	CategoryAssets     FileCategory = "assets"      // Images, icons, fonts, multimedia
	CategoryScripts    FileCategory = "scripts"     // Shell scripts, batch files, one-off deployment runners
	CategoryConfigs    FileCategory = "configs"     // Config files, environment templates, docker definitions
	CategoryTestData   FileCategory = "testdata"    // Test fixtures, mock payloads, sample databases
	CategoryCode       FileCategory = "code"        // Compilable application source code
)

// CodeRole defines the architectural archetype inferred from code AST inspection.
type CodeRole string

const (
	RoleEntry   CodeRole = "entry"   // Application entry points (main functions, CLI commands)
	RoleAPI     CodeRole = "api"     // HTTP handlers, RPC routers, controllers, middleware
	RoleModel   CodeRole = "model"   // Pure entities, data schemas, struct definitions, interfaces
	RoleService CodeRole = "service" // Core business logic, repositories, multi-component orchestrators
	RoleUtil    CodeRole = "util"    // Foundational pure utilities with zero internal imports
	RoleUnknown CodeRole = "unknown" // Unclassified source code
)

// FileNode describes an inspected file in the workspace graph.
type FileNode struct {
	RelPath      string       `json:"rel_path"`
	AbsPath      string       `json:"abs_path"`
	Category     FileCategory `json:"category"`
	Language     string       `json:"language,omitempty"`
	Role         CodeRole     `json:"role,omitempty"`
	PackageName  string       `json:"package_name,omitempty"`
	Imports      []string     `json:"imports,omitempty"`
	ExportedSyms []string     `json:"exported_syms,omitempty"`
	Size         int64        `json:"size"`
	SHA256       string       `json:"sha256"`
	IsRootFile   bool         `json:"is_root_file"`
}

// RelocationOp represents an atomic file move and transformation operation.
type RelocationOp struct {
	SourcePath  string       `json:"source_path"`
	DestPath    string       `json:"dest_path"`
	Category    FileCategory `json:"category"`
	Role        CodeRole     `json:"role"`
	Reason      string       `json:"reason"`
	PackageName string       `json:"new_package_name,omitempty"`
	NeedsImport bool         `json:"needs_import_rewrite"`
}

// AIImpactMetrics quantifies prompt token and context window savings.
type AIImpactMetrics struct {
	RawListingTokensBefore int     `json:"raw_listing_tokens_before"`
	CleanListingTokensAfter int    `json:"clean_listing_tokens_after"`
	TokenSavingsPercent    float64 `json:"token_savings_percent"`
	EstimatedPass1Gain     float64 `json:"estimated_pass1_gain"`
	RepoMapTokensBefore    int     `json:"repo_map_tokens_before"`
	RepoMapTokensAfter     int     `json:"repo_map_tokens_after"`
}

// OrganizationPlan represents the complete layout restructuring plan.
type OrganizationPlan struct {
	WorkspaceRoot string          `json:"workspace_root"`
	GeneratedAt   time.Time       `json:"generated_at"`
	Preset        string          `json:"preset"`
	TotalFiles    int             `json:"total_files"`
	Relocations   []RelocationOp  `json:"relocations"`
	DetectedCycles [][]string     `json:"detected_cycles,omitempty"`
	AIImpact      AIImpactMetrics `json:"ai_impact"`
}

// RollbackOp records a single atomic filesystem change for instant revert.
type RollbackOp struct {
	OriginalPath      string `json:"original_path"`
	CurrentPath       string `json:"current_path"`
	BackupContentPath string `json:"backup_content_path,omitempty"`
	OriginalSHA256    string `json:"original_sha256"`
}

// RollbackJournal records an immutable transaction manifest.
type RollbackJournal struct {
	Timestamp     time.Time    `json:"timestamp"`
	WorkspaceRoot string       `json:"workspace_root"`
	Operations    []RollbackOp `json:"operations"`
	Status        string       `json:"status"` // "COMMITTED", "REVERTED"
}

// EntropyReport captures workspace disorder and token overhead analysis.
type EntropyReport struct {
	WorkspaceRoot        string            `json:"workspace_root"`
	TotalFiles           int               `json:"total_files"`
	RootFilesCount       int               `json:"root_files_count"`
	EntropyScore         float64           `json:"entropy_score"` // 0.0 (pristine) to 1.0 (maximum sprawl)
	CategoryDistribution map[string]int    `json:"category_distribution"`
	RoleDistribution     map[string]int    `json:"role_distribution"`
	TokenOverheadListing int               `json:"token_overhead_listing"`
	DetectedLanguages    []string          `json:"detected_languages"`
	Recommendations      []string          `json:"recommendations"`
}
