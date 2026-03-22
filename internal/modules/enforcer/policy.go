package enforcer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TU_ORG/fenrir/internal/graph"
)

type PolicyEngine struct {
	policies []Policy
	graph   *graph.Graph
}

type Policy struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Pattern     string   `json:"pattern"`
	AllowedIn   []string `json:"allowed_in"`
	ForbiddenIn []string `json:"forbidden_in"`
}

type PolicyConfig struct {
	Team               string   `json:"team"`
	Version            string   `json:"version"`
	ForbiddenLicenses  []string `json:"forbidden_licenses"`
	Policies           []Policy `json:"policies"`
}

type PolicyResult struct {
	Allowed   bool     `json:"allowed"`
	Severity  string   `json:"severity"`
	Violated  []Policy  `json:"violated"`
	Reason    string    `json:"reason"`
}

func NewPolicyEngine(g *graph.Graph) *PolicyEngine {
	return &PolicyEngine{
		graph: g,
	}
}

func (e *PolicyEngine) LoadPolicies(configPath string) error {
	if configPath == "" {
		configPath = ".fenrir/policies.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config PolicyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	e.policies = config.Policies
	return nil
}

func (e *PolicyEngine) CheckAction(action, context string) *PolicyResult {
	result := &PolicyResult{
		Allowed:  true,
		Violated: []Policy{},
	}

	actionLower := strings.ToLower(action)
	contextLower := strings.ToLower(context)

	for _, policy := range e.policies {
		if !e.matchesContext(policy, contextLower) {
			continue
		}

		if e.matchesPattern(policy, actionLower) {
			result.Allowed = false
			result.Severity = policy.Severity
			result.Violated = append(result.Violated, policy)

			switch policy.Severity {
			case "critical":
				result.Reason = "CRITICAL: " + policy.Description
			case "hard":
				result.Reason = "WARNING: " + policy.Description
			case "soft":
				result.Reason = "SUGGESTION: " + policy.Description
			}

			if policy.Severity == "critical" {
				break
			}
		}
	}

	return result
}

func (e *PolicyEngine) matchesContext(policy Policy, context string) bool {
	if len(policy.AllowedIn) == 0 && len(policy.ForbiddenIn) == 0 {
		return true
	}

	for _, allowed := range policy.AllowedIn {
		if strings.Contains(context, allowed) {
			return true
		}
	}

	for _, forbidden := range policy.ForbiddenIn {
		if strings.Contains(context, forbidden) {
			return false
		}
	}

	if len(policy.AllowedIn) > 0 {
		return false
	}

	return true
}

func (e *PolicyEngine) matchesPattern(policy Policy, action string) bool {
	if policy.Pattern == "" {
		return false
	}

	pattern := "(?i)" + policy.Pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return strings.Contains(action, strings.ToLower(policy.Pattern))
	}

	return re.MatchString(action)
}

func (e *PolicyEngine) GetPolicies() []Policy {
	return e.policies
}

func (e *PolicyEngine) AddPolicy(policy Policy) {
	e.policies = append(e.policies, policy)
}

func (e *PolicyEngine) RemovePolicy(policyID string) {
	newPolicies := []Policy{}
	for _, p := range e.policies {
		if p.ID != policyID {
			newPolicies = append(newPolicies, p)
		}
	}
	e.policies = newPolicies
}

func (e *PolicyEngine) SavePolicies(configPath string) error {
	config := PolicyConfig{
		Version:   "1.0",
		Policies:  e.policies,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (e *PolicyEngine) ValidateCode(code, filePath string) []PolicyViolation {
	violations := []PolicyViolation{}

	codeLower := strings.ToLower(code)
	fileLower := strings.ToLower(filePath)

	for _, policy := range e.policies {
		if !e.matchesContext(policy, fileLower) {
			continue
		}

		if e.matchesPattern(policy, codeLower) {
			violations = append(violations, PolicyViolation{
				PolicyID:    policy.ID,
				Description: policy.Description,
				Severity:    policy.Severity,
				File:        filePath,
				Pattern:     policy.Pattern,
			})
		}
	}

	return violations
}

type PolicyViolation struct {
	PolicyID    string `json:"policy_id"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Pattern     string `json:"pattern"`
}
