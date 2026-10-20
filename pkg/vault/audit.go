package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SensitiveFilePatterns lists globs or filenames considered sensitive.
var SensitiveFilePatterns = []string{
	".env",
	".env.local",
	".env.production",
	".env.staging",
	"credentials.json",
	"secrets.json",
	"secrets.yaml",
	"secrets.yml",
}

// AuditWorkspace scans a directory tree for exposed credentials and insecure environment files.
func AuditWorkspace(root string) (*VaultAuditReport, error) {
	var findings []AuditFinding
	filesAudited := 0
	totalSecrets := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if base == ".git" || base == ".taskcache" || base == "node_modules" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		base := strings.ToLower(info.Name())

		isSensitive := false
		if strings.HasPrefix(base, ".env") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".pem") {
			isSensitive = true
		}
		for _, pat := range SensitiveFilePatterns {
			if base == pat {
				isSensitive = true
				break
			}
		}

		if !isSensitive {
			return nil
		}

		filesAudited++
		fFindings, err := auditFile(path, rel)
		if err == nil && len(fFindings) > 0 {
			findings = append(findings, fFindings...)
			totalSecrets += len(fFindings)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	status := "SECURED"
	if totalSecrets > 0 {
		status = "VULNERABLE"
	}

	return &VaultAuditReport{
		WorkspaceRoot:   root,
		FilesAudited:    filesAudited,
		TotalSecrets:    totalSecrets,
		Findings:        findings,
		ProtectedStatus: status,
	}, nil
}

func auditFile(absPath, relPath string) ([]AuditFinding, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var findings []AuditFinding
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		key, val, isCommentOrBlank := ParseEnvLine(line)
		if isCommentOrBlank || key == "" {
			continue
		}

		// Skip if already masked by ZEPHYR vault
		if strings.HasPrefix(val, "ZEPHYR_VAULT_SECRET_") {
			continue
		}

		isSecret, risk := IsSecretVariable(key, val)
		if isSecret {
			findings = append(findings, AuditFinding{
				FilePath:      relPath,
				LineNumber:    lineNum,
				Key:           key,
				Risk:          risk,
				MaskedPreview: MaskSecret(val),
				Description:   fmt.Sprintf("Exposed %s credential on line %d (%s)", risk, lineNum, key),
			})
		}
	}

	return findings, nil
}
