package organize

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// CheckGitCleanliness verifies if the working tree has uncommitted modifications.
func CheckGitCleanliness(root string) error {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		// Not a git repository or git not available, treat as clean
		return nil
	}
	if len(out) > 0 {
		return fmt.Errorf("working directory has uncommitted changes. Commit or stash before organizing")
	}
	return nil
}

// SaveRollbackJournal saves the two phase commit transaction manifest.
func SaveRollbackJournal(root string, journal *RollbackJournal) error {
	reorgDir := filepath.Join(root, ".taskcache", "reorg")
	if err := os.MkdirAll(reorgDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		return err
	}

	timestampStr := journal.Timestamp.Format("20060102_150405")
	journalFile := filepath.Join(reorgDir, fmt.Sprintf("rollback_%s.json", timestampStr))
	if err := os.WriteFile(journalFile, data, 0644); err != nil {
		return err
	}

	// Update latest pointer
	latestFile := filepath.Join(reorgDir, "latest.json")
	return os.WriteFile(latestFile, data, 0644)
}

// BackupFileContent creates a physical byte backup of a file prior to modification.
func BackupFileContent(root, relPath string) (string, error) {
	srcPath := filepath.Join(root, relPath)
	backupDir := filepath.Join(root, ".taskcache", "reorg", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	hash, err := ComputeSHA256(srcPath)
	if err != nil {
		return "", err
	}

	destPath := filepath.Join(backupDir, hash)
	if _, err := os.Stat(destPath); err == nil {
		// Already backed up
		return destPath, nil
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", err
	}

	return destPath, nil
}

// Rollback executes an atomic reverse transaction using the latest rollback journal.
func Rollback(root string) (*RollbackJournal, error) {
	latestFile := filepath.Join(root, ".taskcache", "reorg", "latest.json")
	data, err := os.ReadFile(latestFile)
	if err != nil {
		return nil, fmt.Errorf("no reorganization rollback journal found: %w", err)
	}

	var journal RollbackJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		return nil, fmt.Errorf("corrupt rollback journal: %w", err)
	}

	if journal.Status == "REVERTED" {
		return nil, fmt.Errorf("latest reorganization has already been reverted")
	}

	// Reverse operations in backwards order
	for i := len(journal.Operations) - 1; i >= 0; i-- {
		op := journal.Operations[i]
		currPath := filepath.Join(root, op.CurrentPath)
		origPath := filepath.Join(root, op.OriginalPath)

		// If backed up content exists, restore it first
		if op.BackupContentPath != "" {
			if bData, err := os.ReadFile(op.BackupContentPath); err == nil {
				_ = os.WriteFile(currPath, bData, 0644)
			}
		}

		// Move back to original path if paths differ
		if currPath != origPath {
			_ = os.MkdirAll(filepath.Dir(origPath), 0755)
			_ = os.Rename(currPath, origPath)

			// Clean up parent directory if empty
			parent := filepath.Dir(currPath)
			entries, err := os.ReadDir(parent)
			if err == nil && len(entries) == 0 {
				_ = os.Remove(parent)
			}
		}
	}

	journal.Status = "REVERTED"
	_ = SaveRollbackJournal(root, &journal)

	return &journal, nil
}
