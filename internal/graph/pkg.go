package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type PkgCache struct {
	ID        string    `json:"id"`
	Ecosystem string    `json:"ecosystem"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Exists    bool      `json:"exists"`
	Trusted   bool      `json:"trusted"`
	CVECount  int       `json:"cve_count"`
	License   string    `json:"license"`
	Downloads int       `json:"downloads"`
	AgeDays   int       `json:"age_days"`
	Response  string    `json:"response"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PkgCheckResult struct {
	Exists          bool     `json:"exists"`
	Trusted         bool     `json:"trusted"`
	CVECount        int      `json:"cve_count"`
	License         string   `json:"license"`
	Downloads       int      `json:"downloads"`
	AgeDays         int      `json:"age_days"`
	Warning         string   `json:"warning"`
	SimilarPackages []string `json:"similar_legitimate"`
	CVEs            []CVE    `json:"cves"`
}

type CVE struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
}

type Issue struct {
	Type        string `json:"type"` // vulnerability, license, trust, drift, rule
	Package     string `json:"package,omitempty"`
	Ecosystem   string `json:"ecosystem,omitempty"`
	Version     string `json:"version,omitempty"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

// ─── Validator Module (Fase 2) ───────────────────────────────────────────────

func (g *Graph) CheckPackage(name, ecosystem, version string) (*PkgCheckResult, error) {
	cacheID := fmt.Sprintf("%s:%s:%s", ecosystem, name, version)
	var cached PkgCache

	// Step 1: Check Cache
	err := g.db.QueryRow(`
		SELECT id, ecosystem, name, version, "exists", trusted, cve_count, license, downloads, age_days, response, cached_at, expires_at
		FROM pkg_cache WHERE id = ? AND expires_at > datetime('now')
	`, cacheID).Scan(&cached.ID, &cached.Ecosystem, &cached.Name, &cached.Version, &cached.Exists, &cached.Trusted, &cached.CVECount, &cached.License, &cached.Downloads, &cached.AgeDays, &cached.Response, &cached.CachedAt, &cached.ExpiresAt)

	if err == nil {
		var result PkgCheckResult
		if cached.Response != "" {
			json.Unmarshal([]byte(cached.Response), &result)
			return &result, nil
		}
		return &PkgCheckResult{
			Exists:    cached.Exists,
			Trusted:   cached.Trusted,
			CVECount:  cached.CVECount,
			License:   cached.License,
			Downloads: cached.Downloads,
			AgeDays:   cached.AgeDays,
		}, nil
	}

	// Step 2: Live Check (Real integration with APIs)
	result := &PkgCheckResult{
		Exists:  true,
		Trusted: true,
	}

	// Check for CVEs via OSV.dev
	cves, err := g.fetchCVEs(name, ecosystem, version)
	if err == nil {
		result.CVEs = cves
		result.CVECount = len(cves)
		if len(cves) > 0 {
			result.Trusted = false
			result.Warning = fmt.Sprintf("Found %d known vulnerabilities", len(cves))
		}
	}
	// Check License
	license, _ := g.fetchLicense(name, ecosystem, version)
	result.License = license
	if !g.checkLicenseCompatibility(license) {
		result.Trusted = false
		result.Warning = fmt.Sprintf("Forbidden license detected: %s", license)
	}

	// Check Typosquatting (Simple local heuristic for now)
	if isSuspiciousName(name) {
		result.Warning = "Package name looks suspicious (possible typosquatting)"
		result.Trusted = false
	}

	// Step 3: Persistence
	responseJSON, _ := json.Marshal(result)
	expiresAt := time.Now().Add(24 * time.Hour) // Cache for 24h
	_, err = g.db.Exec(`
		INSERT INTO pkg_cache (id, ecosystem, name, version, "exists", trusted, cve_count, license, downloads, age_days, response, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			"exists"=excluded."exists",
			trusted=excluded.trusted,
			cve_count=excluded.cve_count,
			response=excluded.response,
			expires_at=excluded.expires_at
	`, cacheID, ecosystem, name, version, result.Exists, result.Trusted, result.CVECount, result.License, result.Downloads, result.AgeDays, string(responseJSON), expiresAt)

	return result, err
}

func (g *Graph) AuditDependencies(manifestPath string) ([]Issue, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}

	issues := []Issue{}

	// TODO: Support more manifest types (requirements.txt, Cargo.toml, etc.)
	if strings.HasSuffix(manifestPath, "package.json") {
		var pkgJSON struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if err := json.Unmarshal(data, &pkgJSON); err != nil {
			return nil, fmt.Errorf("parse package.json: %w", err)
		}

		for name, version := range pkgJSON.Dependencies {
			i, _ := g.auditPkg(name, "npm", version)
			issues = append(issues, i...)
		}
		for name, version := range pkgJSON.DevDependencies {
			i, _ := g.auditPkg(name, "npm", version)
			issues = append(issues, i...)
		}
	}

	return issues, nil
}

func (g *Graph) auditPkg(name, ecosystem, version string) ([]Issue, error) {
	cleanVer := strings.TrimPrefix(version, "^")
	cleanVer = strings.TrimPrefix(cleanVer, "~")
	
	res, err := g.CheckPackage(name, ecosystem, cleanVer)
	if err != nil {
		return nil, err
	}

	issues := []Issue{}
	if res.CVECount > 0 {
		issues = append(issues, Issue{
			Type:        "vulnerability",
			Package:     name,
			Ecosystem:   ecosystem,
			Version:     version,
			Severity:    "high",
			Message:     fmt.Sprintf("Found %d vulnerabilities", res.CVECount),
			Description: fmt.Sprintf("Found %d known vulnerabilities at osv.dev", res.CVECount),
		})
	}
	if !res.Trusted {
		issues = append(issues, Issue{
			Type:        "trust",
			Package:     name,
			Ecosystem:   ecosystem,
			Version:     version,
			Severity:    "medium",
			Message:     "Package fails trust checks (e.g. suspicious name)",
			Description: "Package fails trust checks (e.g. suspicious name or metadata)",
		})
	}
	return issues, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (g *Graph) isOffline() bool {
	return os.Getenv("FENRIR_OFFLINE") == "true"
}

func (g *Graph) postWithRetry(url string, body []byte) (*http.Response, error) {
	if g.isOffline() {
		return nil, fmt.Errorf("fenrir is in offline mode")
	}

	var resp *http.Response
	var err error
	backoff := 500 * time.Millisecond

	for i := 0; i < 3; i++ {
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(body))
		if err == nil {
			if resp.StatusCode == 429 {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return resp, nil
		}
		time.Sleep(backoff)
		backoff *= 2
	}
	return resp, err
}

func (g *Graph) fetchCVEs(name, ecosystem, version string) ([]CVE, error) {
	apiEcosystem := mapEcosystemToOSV(ecosystem)
	if apiEcosystem == "" {
		return nil, nil
	}

	query := map[string]interface{}{
		"package": map[string]string{
			"name":      name,
			"ecosystem": apiEcosystem,
		},
	}
	if version != "" {
		query["version"] = version
	}

	body, _ := json.Marshal(query)
	resp, err := g.postWithRetry("https://api.osv.dev/v1/query", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSV API error: %d", resp.StatusCode)
	}

	var osvResponse struct {
		Vulns []struct {
			ID      string `json:"id"`
			Summary string `json:"summary"`
		} `json:"vulns"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&osvResponse); err != nil {
		return nil, err
	}

	cves := make([]CVE, 0, len(osvResponse.Vulns))
	for _, v := range osvResponse.Vulns {
		cves = append(cves, CVE{
			ID:      v.ID,
			Summary: v.Summary,
			Severity: "medium", // OSV doesn't give a single severity field easily
		})
	}

	return cves, nil
}

func (g *Graph) fetchLicense(name, ecosystem, version string) (string, error) {
	if g.isOffline() {
		return "UNKNOWN", nil
	}

	url := fmt.Sprintf("https://api.deps.dev/v3/systems/%s/packages/%s/versions/%s", ecosystem, name, version)
	resp, err := http.Get(url)
	if err != nil {
		return "UNKNOWN", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "UNKNOWN", nil
	}

	var depsResp struct {
		Licenses []string `json:"licenses"`
	}
	json.NewDecoder(resp.Body).Decode(&depsResp)

	if len(depsResp.Licenses) > 0 {
		return depsResp.Licenses[0], nil
	}
	return "UNKNOWN", nil
}

func (g *Graph) checkLicenseCompatibility(license string) bool {
	forbidden := []string{"AGPL", "GPL", "LGPL", "BSL"}
	for _, f := range forbidden {
		if strings.Contains(strings.ToUpper(license), f) {
			return false
		}
	}
	return true
}

func mapEcosystemToOSV(e string) string {
	switch strings.ToLower(e) {
	case "npm":
		return "npm"
	case "pypi":
		return "PyPI"
	case "cargo":
		return "crates.io"
	case "nuget":
		return "NuGet"
	default:
		return ""
	}
}

func isSuspiciousName(name string) bool {
	// Simple examples of typosquatting patterns:
	// - Very similar to popular packages (lodash vs lodaash) - would need a more complex list
	// - Contains common suffixes like -security, -helper when not expected
	suspiciousSuffixes := []string{"-security", "-fix", "-hotfix", "-patch"}
	for _, s := range suspiciousSuffixes {
		if strings.HasSuffix(name, s) {
			return true
		}
	}
	return false
}
