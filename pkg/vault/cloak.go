package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseEnvLine decomposes a single environment file line into key and value.
func ParseEnvLine(line string) (key, val string, isCommentOrBlank bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", true
	}

	idx := strings.Index(line, "=")
	if idx == -1 {
		return strings.TrimSpace(line), "", false
	}

	key = strings.TrimSpace(line[:idx])
	val = strings.TrimSpace(line[idx+1:])

	// Strip surrounding quotes if present
	if (strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`)) ||
		(strings.HasPrefix(val, `'`) && strings.HasSuffix(val, `'`)) {
		if len(val) >= 2 {
			val = val[1 : len(val)-1]
		}
	}

	return key, val, false
}

// CloakEnvContent redacts all sensitive credential values from environment text.
func CloakEnvContent(filePath string, rawContent string) (*CloakResult, error) {
	scanner := bufio.NewScanner(strings.NewReader(rawContent))
	var outLines []string
	secretMap := make(map[string]string)

	totalVars := 0
	secretCount := 0
	publicCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		key, val, isCommentOrBlank := ParseEnvLine(line)

		if isCommentOrBlank {
			outLines = append(outLines, line)
			continue
		}

		totalVars++
		isSecret, _ := IsSecretVariable(key, val)

		if isSecret {
			secretCount++
			placeholder := GenerateRedactedToken(key, val)
			secretMap[placeholder] = val
			outLines = append(outLines, fmt.Sprintf("%s=%s", key, placeholder))
		} else {
			publicCount++
			outLines = append(outLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &CloakResult{
		OriginalPath:   filePath,
		CloakedContent: strings.Join(outLines, "\n"),
		TotalVariables: totalVars,
		SecretCount:    secretCount,
		PublicCount:    publicCount,
		SecretMap:      secretMap,
	}, nil
}

// CloakFile reads a file from disk and returns the AI-safe cloaked representation.
func CloakFile(filePath string) (*CloakResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return CloakEnvContent(filePath, string(data))
}
