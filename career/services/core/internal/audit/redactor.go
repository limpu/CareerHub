package audit

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	bearerTokenRegex = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+\/]+=*`)
	passwordJsonRegex = regexp.MustCompile(`(?i)"(password|secret|token|refresh_token|api_key)"\s*:\s*"[^"]+"`)
	passwordUrlRegex  = regexp.MustCompile(`(?i)(password|token|secret)=[^&\s]+`)
	ssnRegex          = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	resumeTextRegex   = regexp.MustCompile(`(?i)(raw_resume_text|resume_blob|curriculum vitae)[\s\S]{30,}`)
)

// RedactText sanitizes log lines and error strings, masking any secrets or raw resumes (REQ-023).
func RedactText(input string) string {
	if input == "" {
		return ""
	}

	sanitized := bearerTokenRegex.ReplaceAllString(input, "Bearer [REDACTED_TOKEN]")
	sanitized = passwordJsonRegex.ReplaceAllStringFunc(sanitized, func(match string) string {
		parts := strings.Split(match, ":")
		if len(parts) == 2 {
			return parts[0] + `: "[REDACTED_SECRET]"`
		}
		return `"[REDACTED_SECRET]"`
	})
	sanitized = passwordUrlRegex.ReplaceAllString(sanitized, "$1=[REDACTED_SECRET]")
	sanitized = ssnRegex.ReplaceAllString(sanitized, "[REDACTED_SSN]")
	sanitized = resumeTextRegex.ReplaceAllString(sanitized, "[REDACTED_RAW_RESUME_CONTENT]")

	return sanitized
}

// RedactMap deeply scrubs a key-value map before saving to audit storage or returning in diagnostics.
func RedactMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}

	cleaned := make(map[string]interface{}, len(input))
	for k, v := range input {
		lowerK := strings.ToLower(k)
		if strings.Contains(lowerK, "password") ||
			strings.Contains(lowerK, "token") ||
			strings.Contains(lowerK, "secret") ||
			strings.Contains(lowerK, "key") ||
			strings.Contains(lowerK, "credential") {
			cleaned[k] = "[REDACTED_SECRET]"
			continue
		}

		if strings.Contains(lowerK, "raw_resume") ||
			strings.Contains(lowerK, "resume_blob") ||
			strings.Contains(lowerK, "full_inbox") {
			cleaned[k] = "[REDACTED_RESUME_DATA]"
			continue
		}

		switch val := v.(type) {
		case string:
			cleaned[k] = RedactText(val)
		case map[string]interface{}:
			cleaned[k] = RedactMap(val)
		default:
			cleaned[k] = fmt.Sprintf("%v", val)
		}
	}

	return cleaned
}
