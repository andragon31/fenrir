package shield

import (
	"regexp"
	"strings"
)

type Sanitizer struct {
	secretPatterns []*SecretPattern
	injectPatterns []*InjectPattern
}

type SecretPattern struct {
	Name    string
	Pattern *regexp.Regexp
	Replace string
}

type InjectPattern struct {
	Pattern     string
	Severity    string
	Description string
}

func NewSanitizer() *Sanitizer {
	s := &Sanitizer{
		secretPatterns: []*SecretPattern{
			{
				Name:    "OpenAI API Key",
				Pattern: regexp.MustCompile(`(?i)(sk-[A-Za-z0-9_-]{20,})`),
				Replace: "[OPENAI_KEY_REDACTED]",
			},
			{
				Name:    "Anthropic API Key",
				Pattern: regexp.MustCompile(`(?i)(sk-ant-[A-Za-z0-9_-]{20,})`),
				Replace: "[ANTHROPIC_KEY_REDACTED]",
			},
			{
				Name:    "GitHub Token",
				Pattern: regexp.MustCompile(`(?i)(ghp_[A-Za-z0-9]{36}|github_pat_[A-Za-z0-9_]{22,})`),
				Replace: "[GITHUB_TOKEN_REDACTED]",
			},
			{
				Name:    "AWS Access Key",
				Pattern: regexp.MustCompile(`(?i)(AKIA[A-Z0-9]{16})`),
				Replace: "[AWS_KEY_REDACTED]",
			},
			{
				Name:    "AWS Secret Key",
				Pattern: regexp.MustCompile(`(?i)(aws_secret_access_key[=:\s]*['\"]?[A-Za-z0-9/+=]{40})`),
				Replace: "[AWS_SECRET_REDACTED]",
			},
			{
				Name:    "JWT Token",
				Pattern: regexp.MustCompile(`eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+`),
				Replace: "[JWT_TOKEN_REDACTED]",
			},
			{
				Name:    "Google API Key",
				Pattern: regexp.MustCompile(`(?i)(AIza[A-Za-z0-9_-]{35})`),
				Replace: "[GOOGLE_KEY_REDACTED]",
			},
			{
				Name:    "Stripe API Key",
				Pattern: regexp.MustCompile(`(?i)(sk_live_[A-Za-z0-9]{24,})`),
				Replace: "[STRIPE_KEY_REDACTED]",
			},
			{
				Name:    "Private Key",
				Pattern: regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
				Replace: "[PRIVATE_KEY_REDACTED]",
			},
			{
				Name:    "Generic API Key",
				Pattern: regexp.MustCompile(`(?i)(api[_-]?key[=:\s]*['\"]?[A-Za-z0-9_-]{20,})`),
				Replace: "[API_KEY_REDACTED]",
			},
			{
				Name:    "Password in URL",
				Pattern: regexp.MustCompile(`(?i)(://[^:]+:)[^@]+(@)`),
				Replace: "${1}[PASSWORD_REDACTED]${2}",
			},
			{
				Name:    "Bearer Token",
				Pattern: regexp.MustCompile(`(?i)(Bearer\s+[A-Za-z0-9_-]{20,})`),
				Replace: "Bearer [TOKEN_REDACTED]",
			},
			{
				Name:    "Basic Auth",
				Pattern: regexp.MustCompile(`(?i)(Basic\s+[A-Za-z0-9+/=]{20,})`),
				Replace: "Basic [AUTH_REDACTED]",
			},
		},
		injectPatterns: []*InjectPattern{
			{
				Pattern:     "(?i)ignore\\s+(all\\s+)?previous\\s+(instructions?|commands?|rules?)",
				Severity:    "high",
				Description: "Attempt to ignore previous instructions",
			},
			{
				Pattern:     "(?i)disregard\\s+(all\\s+)?(your\\s+)?(instructions?|rules?|guidelines?)",
				Severity:    "high",
				Description: "Attempt to disregard system rules",
			},
			{
				Pattern:     "(?i)forget\\s+(everything|all\\s+previous|what\\s+you\\s+know)",
				Severity:    "high",
				Description: "Attempt to reset memory",
			},
			{
				Pattern:     "(?i)you\\s+are\\s+now\\s+(a|an)\\s+",
				Severity:    "medium",
				Description: "Attempt to role-play or persona injection",
			},
			{
				Pattern:     "(?i)pretend\\s+(you\\s+are|to\\s+be)",
				Severity:    "medium",
				Description: "Attempt to change identity",
			},
			{
				Pattern:     "(?i)new\\s+(system|initial)\\s+instructions?",
				Severity:    "high",
				Description: "Attempt to inject new system prompt",
			},
			{
				Pattern:     "(?i)reveal\\s+(your|the)\\s+(system\\s+)?prompt",
				Severity:    "medium",
				Description: "Attempt to extract system prompt",
			},
			{
				Pattern:     "(?i)sudo\\s+rm\\s+-rf",
				Severity:    "critical",
				Description: "Destructive command pattern",
			},
			{
				Pattern:     "(?i)<!--\\s*",
				Severity:    "medium",
				Description: "HTML comment injection attempt",
			},
			{
				Pattern:     "(?i)\\{{2,}.*\\}{2,}",
				Severity:    "medium",
				Description: "Template injection attempt",
			},
			{
				Pattern:     "(?i)\\$\\(.*\\)",
				Severity:    "low",
				Description: "Command substitution attempt",
			},
			{
				Pattern:     "(?i)`.*`",
				Severity:    "low",
				Description: "Backtick command execution attempt",
			},
		},
	}

	return s
}

func (s *Sanitizer) SanitizeSecrets(content string) string {
	result := content

	for _, pattern := range s.secretPatterns {
		result = pattern.Pattern.ReplaceAllString(result, pattern.Replace)
	}

	return result
}

func (s *Sanitizer) StripPrivateTags(content string) string {
	replacer := strings.NewReplacer(
		"<private>", "[REDACTED]",
		"</private>", "[/REDACTED]",
		"<PRIVATE>", "[REDACTED]",
		"</PRIVATE>", "[/REDACTED]",
	)

	return replacer.Replace(content)
}

func (s *Sanitizer) CheckInjection(content string) (*InjectionResult, error) {
	result := &InjectionResult{
		Clean:       true,
		Detected:    []InjectDetection{},
		TotalCount:  0,
		HighSeverity: 0,
	}

	contentLower := strings.ToLower(content)

	for _, pattern := range s.injectPatterns {
		if pattern.Pattern != "" {
			re := regexp.MustCompile(pattern.Pattern)
			matches := re.FindAllStringIndex(contentLower, -1)
			
			if len(matches) > 0 {
				result.Clean = false
				result.TotalCount += len(matches)

				if pattern.Severity == "high" || pattern.Severity == "critical" {
					result.HighSeverity += len(matches)
				}

				result.Detected = append(result.Detected, InjectDetection{
					Pattern:     pattern.Pattern,
					Severity:    pattern.Severity,
					Description: pattern.Description,
					Count:       len(matches),
				})
			}
		}
	}

	return result, nil
}

type InjectionResult struct {
	Clean        bool             `json:"clean"`
	Detected     []InjectDetection `json:"detected"`
	TotalCount   int              `json:"total_count"`
	HighSeverity int              `json:"high_severity"`
}

type InjectDetection struct {
	Pattern     string `json:"pattern"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

func (s *Sanitizer) SanitizeAll(content string) string {
	content = s.StripPrivateTags(content)
	content = s.SanitizeSecrets(content)
	return content
}
