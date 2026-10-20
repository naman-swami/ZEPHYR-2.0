package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var (
	// Known public configuration keys that are safe to expose to AI agents
	publicConfigKeys = map[string]bool{
		"PORT":            true,
		"HOST":            true,
		"ENV":             true,
		"ENVIRONMENT":     true,
		"DEBUG":           true,
		"LOG_LEVEL":       true,
		"LOG_FORMAT":      true,
		"NODE_ENV":        true,
		"APP_NAME":        true,
		"APP_ENV":         true,
		"VERSION":         true,
		"TZ":              true,
		"TIMEZONE":        true,
		"TIMEOUT":         true,
		"WORKERS":         true,
		"REGION":          true,
		"AWS_REGION":      true,
		"AWS_DEFAULT_REGION": true,
		"DATABASE_DRIVER": true,
		"DB_HOST":         true,
		"DB_PORT":         true,
		"DB_NAME":         true,
		"REDIS_HOST":      true,
		"REDIS_PORT":      true,
	}

	// Secret key name substrings
	secretKeyKeywords = []string{
		"KEY",
		"SECRET",
		"PASSWORD",
		"PASSWD",
		"PASS",
		"TOKEN",
		"AUTH",
		"PRIVATE",
		"CREDENTIAL",
		"SALT",
		"SIGNING",
		"CERT",
		"BEARER",
		"DSSE",
		"ACCESS_KEY",
		"APIKEY",
	}

	// High confidence secret value patterns
	openAIKeyRegex    = regexp.MustCompile(`sk-[a-zA-Z0-9_-]{20,}`)
	githubTokenRegex  = regexp.MustCompile(`gh[pousr]_[a-zA-Z0-9]{20,}`)
	awsAccessKeyRegex = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	jwtTokenRegex     = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{10,}\.eyJ[a-zA-Z0-9_-]{10,}`)
	connStringRegex   = regexp.MustCompile(`(?i)(?:postgres|postgresql|mysql|mongodb|redis):\/\/[^:\s]+:([^@\s]+)@`)
)

// IsSecretVariable evaluates if a key and value represent confidential credentials.
func IsSecretVariable(key, value string) (bool, RiskLevel) {
	upperKey := strings.ToUpper(strings.TrimSpace(key))
	cleanVal := strings.TrimSpace(value)

	// Explicit public config keys are never secrets unless value matches a known token pattern
	if publicConfigKeys[upperKey] && !matchesSecretPattern(cleanVal) {
		return false, RiskLow
	}

	// 1. Check value patterns (Critical risk)
	if openAIKeyRegex.MatchString(cleanVal) || githubTokenRegex.MatchString(cleanVal) || awsAccessKeyRegex.MatchString(cleanVal) {
		return true, RiskCritical
	}
	if jwtTokenRegex.MatchString(cleanVal) || connStringRegex.MatchString(cleanVal) {
		return true, RiskCritical
	}

	// 2. Check key keywords
	for _, kw := range secretKeyKeywords {
		if strings.Contains(upperKey, kw) {
			if len(cleanVal) > 0 && cleanVal != `""` && cleanVal != "''" {
				return true, RiskHigh
			}
		}
	}

	// 3. High entropy value heuristics: string length >= 32 with zero spaces
	if len(cleanVal) >= 32 && !strings.Contains(cleanVal, " ") && !strings.HasPrefix(cleanVal, "http") {
		return true, RiskMedium
	}

	return false, RiskLow
}

func matchesSecretPattern(val string) bool {
	return openAIKeyRegex.MatchString(val) || githubTokenRegex.MatchString(val) || awsAccessKeyRegex.MatchString(val)
}

// GenerateRedactedToken creates a deterministic synthetic placeholder for a secret.
func GenerateRedactedToken(key, value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(key + ":" + value))
	sumHex := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("ZEPHYR_VAULT_SECRET_%s", sumHex[:8])
}

// MaskSecret returns a safe terminal preview masking the majority of characters.
func MaskSecret(val string) string {
	clean := strings.TrimSpace(val)
	if len(clean) <= 6 {
		return "******"
	}
	return clean[:3] + "..." + clean[len(clean)-3:]
}
