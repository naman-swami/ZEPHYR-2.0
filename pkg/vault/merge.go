package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// MergeEnvWithAI reconciles an AI proposed environment file with authentic user credentials.
func MergeEnvWithAI(targetFile, existingContent, proposedContent string) (*MergeResult, error) {
	// 1. Build lookup tables from existing authentic environment file
	existingPairs := make(map[string]string)
	placeholderMap := make(map[string]string)
	existingScanner := bufio.NewScanner(strings.NewReader(existingContent))

	for existingScanner.Scan() {
		line := existingScanner.Text()
		key, val, isCommentOrBlank := ParseEnvLine(line)
		if !isCommentOrBlank && key != "" {
			existingPairs[key] = val
			isSec, _ := IsSecretVariable(key, val)
			if isSec {
				token := GenerateRedactedToken(key, val)
				placeholderMap[token] = val
			}
		}
	}

	// 2. Parse AI proposed content and reconcile
	proposedScanner := bufio.NewScanner(strings.NewReader(proposedContent))
	var mergedLines []string
	seenKeys := make(map[string]bool)

	preservedSecrets := 0
	var addedVars []string
	var updatedVars []string

	for proposedScanner.Scan() {
		line := proposedScanner.Text()
		key, val, isCommentOrBlank := ParseEnvLine(line)

		if isCommentOrBlank {
			mergedLines = append(mergedLines, line)
			continue
		}

		seenKeys[key] = true
		origVal, hadOriginal := existingPairs[key]
		origIsSecret, _ := IsSecretVariable(key, origVal)

		// Case A: AI passed a ZEPHYR synthetic placeholder token
		if strings.HasPrefix(val, "ZEPHYR_VAULT_SECRET_") {
			if authentic, found := placeholderMap[val]; found {
				mergedLines = append(mergedLines, fmt.Sprintf("%s=%s", key, authentic))
				preservedSecrets++
				continue
			}
			if hadOriginal && origIsSecret {
				mergedLines = append(mergedLines, fmt.Sprintf("%s=%s", key, origVal))
				preservedSecrets++
				continue
			}
		}

		// Case B: Existing secret key where AI supplied dummy text or template placeholder
		if hadOriginal && origIsSecret && isDummyOrPlaceholder(val) {
			mergedLines = append(mergedLines, fmt.Sprintf("%s=%s", key, origVal))
			preservedSecrets++
			continue
		}

		// Case C: Brand new variable introduced by AI
		if !hadOriginal {
			addedVars = append(addedVars, key)
			mergedLines = append(mergedLines, line)
			continue
		}

		// Case D: Existing public variable modified by AI
		if hadOriginal && origVal != val {
			updatedVars = append(updatedVars, key)
			mergedLines = append(mergedLines, line)
			continue
		}

		// Case E: Unchanged variable
		mergedLines = append(mergedLines, line)
	}

	if err := proposedScanner.Err(); err != nil {
		return nil, err
	}

	mergedText := strings.Join(mergedLines, "\n")

	return &MergeResult{
		TargetFile:       targetFile,
		Timestamp:        time.Now(),
		PreservedSecrets: preservedSecrets,
		AddedVariables:   addedVars,
		UpdatedVariables: updatedVars,
		MergedContent:    mergedText,
	}, nil
}

// isDummyOrPlaceholder checks if a value looks like an AI hallucinated template value.
func isDummyOrPlaceholder(val string) bool {
	lower := strings.ToLower(val)
	return strings.Contains(lower, "your_") ||
		strings.Contains(lower, "replace_me") ||
		strings.Contains(lower, "insert_here") ||
		strings.Contains(lower, "dummy") ||
		strings.Contains(lower, "placeholder") ||
		val == `""` || val == `''` || val == ""
}

// ApplyMergeToFile safely writes the reconciled environment text to disk.
func ApplyMergeToFile(targetFile, existingContent, proposedContent string) (*MergeResult, error) {
	res, err := MergeEnvWithAI(targetFile, existingContent, proposedContent)
	if err != nil {
		return nil, err
	}

	// Atomic write
	tmpFile := targetFile + ".zephyr_vault_tmp"
	if err := os.WriteFile(tmpFile, []byte(res.MergedContent), 0600); err != nil {
		return nil, err
	}

	if err := os.Rename(tmpFile, targetFile); err != nil {
		_ = os.Remove(tmpFile)
		return nil, err
	}

	return res, nil
}
