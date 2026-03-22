package validator

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/andragon31/fenrir/internal/graph"
)

type RegistryValidator struct {
	httpClient *http.Client
	cache      *CacheValidator
}

type PackageInfo struct {
	Name             string `json:"name"`
	Exists           bool   `json:"exists"`
	Trusted          bool   `json:"trusted"`
	CVECount         int    `json:"cve_count"`
	License          string `json:"license"`
	DownloadsMonthly int    `json:"downloads_monthly"`
	AgeDays          int    `json:"age_days"`
	Version          string `json:"version"`
	SimilarPackages  []string `json:"similar_legitimate"`
	CVEs             []CVE  `json:"cves"`
	Warning          string `json:"warning"`
}

type CVE struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
}

type LicenseCheck struct {
	License  string `json:"license"`
	Allowed  bool   `json:"allowed"`
	Category string `json:"category"`
}

var (
	forbiddenLicenses = []string{
		"GPL-3.0", "AGPL-3.0", "GPL-2.0", "LGPL-2.1",
		"SSPL-1.0", "Elastic-2.0", "Commons-Clause",
	}

	allowedLicenses = []string{
		"MIT", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause",
		"ISC", "Unlicense", "MPL-2.0", "0BSD",
	}

	typosquattingPatterns = []struct {
		original  string
		variants  []string
	}{
		{"lodash", {"1odash", "lodash-", "-lodash", "loddash", "lodasht"}},
		{"request", {"requeset", "reqest", "request-", "-request"}},
		{"express", {"expres", "expresjs", "express-", "exprss"}},
		{"axios", {"axois", "aixoz", "axios-", "-axios"}},
		{"moment", {"momemt", "momnet", "moment-", "mmoment"}},
		{"react", {"reacet", "reackt", "react-", "-react"}},
		{"vue", {"vu", "vuejs", "vue-"}},
		{"webpack", {"webpak", "webpack-", "webpck"}},
		{"eslint", {"eslint", "eslint-", "-eslint"}},
		{"typescript", {"typescrip", "typescrypt", "tscript"}},
	}
)

func NewRegistryValidator() *RegistryValidator {
	return &RegistryValidator{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
		},
		cache: NewCacheValidator(),
	}
}

func (v *RegistryValidator) CheckNPM(pkgName, version string) (*PackageInfo, error) {
	cacheKey := fmt.Sprintf("npm:%s:%s", pkgName, version)
	if cached := v.cache.Get(cacheKey); cached != nil {
		return cached, nil
	}

	url := fmt.Sprintf("https://registry.npmjs.org/%s", pkgName)
	if version != "" {
		url = fmt.Sprintf("https://registry.npmjs.org/%s/%s", pkgName, version)
	}

	resp, err := v.httpClient.Get(url)
	if err != nil {
		return &PackageInfo{
			Name:    pkgName,
			Exists:  false,
			Trusted: false,
			Warning: "Unable to verify package - offline mode",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		info := &PackageInfo{
			Name:    pkgName,
			Exists:  false,
			Trusted: false,
			Warning: "Package not found in npm registry",
		}
		v.cache.Set(cacheKey, info)
		return info, nil
	}

	body, _ := io.ReadAll(resp.Body)
	var npmResp struct {
		Name        string `json:"name"`
		Dist        struct {
			Tarball string `json:"tarball"`
		} `json:"dist"`
		Time        string `json:"time"`
		Downloads   int    `json:"downloads"`
		License     string `json:"license"`
		Version     string `json:"version"`
	}

	json.Unmarshal(body, &npmResp)

	info := &PackageInfo{
		Name:             npmResp.Name,
		Exists:           true,
		Trusted:          true,
		Version:          npmResp.Version,
		License:          npmResp.License,
		DownloadsMonthly: npmResp.Downloads,
		AgeDays:          v.calculateAge(npmResp.Time),
	}

	if v.isSuspicious(pkgName) {
		info.Trusted = false
		info.Warning = "SUSPICIOUS: Similar to popular package (possible typosquatting)"
	}

	v.cache.Set(cacheKey, info)
	return info, nil
}

func (v *RegistryValidator) CheckPyPI(pkgName, version string) (*PackageInfo, error) {
	cacheKey := fmt.Sprintf("pypi:%s:%s", pkgName, version)
	if cached := v.cache.Get(cacheKey); cached != nil {
		return cached, nil
	}

	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", pkgName)
	if version != "" {
		url = fmt.Sprintf("https://pypi.org/pypi/%s/%s/json", pkgName, version)
	}

	resp, err := v.httpClient.Get(url)
	if err != nil {
		return &PackageInfo{
			Name:    pkgName,
			Exists:  false,
			Trusted: false,
			Warning: "Unable to verify package - offline mode",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		info := &PackageInfo{
			Name:    pkgName,
			Exists:  false,
			Trusted: false,
			Warning: "Package not found in PyPI",
		}
		v.cache.Set(cacheKey, info)
		return info, nil
	}

	body, _ := io.ReadAll(resp.Body)
	var pypiResp struct {
		Info struct {
			Name            string `json:"name"`
			License         string `json:"license"`
			Version         string `json:"version"`
			Summary         string `json:"summary"`
			DownloadsLast30 int    `json:"downloads_last_30_days"`
		} `json:"info"`
		Urls []struct {
			UploadTime string `json:"upload_time"`
		} `json:"urls"`
	}

	json.Unmarshal(body, &pypiResp)

	info := &PackageInfo{
		Name:             pypiResp.Info.Name,
		Exists:           true,
		Trusted:          true,
		Version:          pypiResp.Info.Version,
		License:          pypiResp.Info.License,
		DownloadsMonthly: pypiResp.Info.DownloadsLast30,
	}

	if len(pypiResp.Urls) > 0 {
		info.AgeDays = v.calculateAge(pypiResp.Urls[0].UploadTime)
	}

	if v.isSuspicious(pkgName) {
		info.Trusted = false
		info.Warning = "SUSPICIOUS: Similar to popular package (possible typosquatting)"
	}

	v.cache.Set(cacheKey, info)
	return info, nil
}

func (v *RegistryValidator) CheckCratesIO(pkgName, version string) (*PackageInfo, error) {
	cacheKey := fmt.Sprintf("cargo:%s:%s", pkgName, version)
	if cached := v.cache.Get(cacheKey); cached != nil {
		return cached, nil
	}

	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", pkgName)
	if version != "" {
		url = fmt.Sprintf("https://crates.io/api/v1/crates/%s/%s", pkgName, version)
	}

	resp, err := v.httpClient.Get(url)
	if err != nil {
		return &PackageInfo{
			Name:    pkgName,
			Exists:  false,
			Trusted: false,
			Warning: "Unable to verify package - offline mode",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		info := &PackageInfo{
			Name:    pkgName,
			Exists:  false,
			Trusted: false,
			Warning: "Package not found in crates.io",
		}
		v.cache.Set(cacheKey, info)
		return info, nil
	}

	body, _ := io.ReadAll(resp.Body)
	var cratesResp struct {
		Crate struct {
			Name          string `json:"name"`
			License       string `json:"license"`
			Downloads     int    `json:"downloads"`
			RecentDownloads int  `json:"recent_downloads"`
			MaxVersion    string `json:"max_version"`
			CreatedAt     string `json:"created_at"`
		} `json:"crate"`
	}

	json.Unmarshal(body, &cratesResp)

	info := &PackageInfo{
		Name:             cratesResp.Crate.Name,
		Exists:           true,
		Trusted:          true,
		Version:          cratesResp.Crate.MaxVersion,
		License:          cratesResp.Crate.License,
		DownloadsMonthly: cratesResp.Crate.RecentDownloads,
		AgeDays:          v.calculateAge(cratesResp.Crate.CreatedAt),
	}

	v.cache.Set(cacheKey, info)
	return info, nil
}

func (v *RegistryValidator) CheckNuGet(pkgName, version string) (*PackageInfo, error) {
	cacheKey := fmt.Sprintf("nuget:%s:%s", pkgName, version)
	if cached := v.cache.Get(cacheKey); cached != nil {
		return cached, nil
	}

	url := fmt.Sprintf("https://api.nuget.org/v3/registration5-semver1/%s/index.json", strings.ToLower(pkgName))

	resp, err := v.httpClient.Get(url)
	if err != nil {
		return &PackageInfo{
			Name:    pkgName,
			Exists:  false,
			Trusted: false,
			Warning: "Unable to verify package - offline mode",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		info := &PackageInfo{
			Name:    pkgName,
			Exists:  false,
			Trusted: false,
			Warning: "Package not found in NuGet",
		}
		v.cache.Set(cacheKey, info)
		return info, nil
	}

	info := &PackageInfo{
		Name:    pkgName,
		Exists:  true,
		Trusted: true,
		License: "NuGet",
	}

	v.cache.Set(cacheKey, info)
	return info, nil
}

func (v *RegistryValidator) CheckOSV(pkgName, ecosystem string) ([]CVE, error) {
	query := map[string]interface{}{
		"package": map[string]string{
			"name":      pkgName,
			"ecosystem": ecosystem,
		},
	}

	body, _ := json.Marshal(query)
	req, _ := http.NewRequest("POST", "https://api.osv.dev/v1/query", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return []CVE{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []CVE{}, nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	var osvResp struct {
		Vulns []struct {
			ID       string `json:"id"`
			Severity struct {
				Score string `json:"score"`
			} `json:"severity"`
			Summary string `json:"summary"`
		} `json:"vulns"`
	}

	json.Unmarshal(respBody, &osvResp)

	var cves []CVE
	for _, vuln := range osvResp.Vulns {
		cves = append(cves, CVE{
			ID:       vuln.ID,
			Severity: vuln.Severity.Score,
			Summary:  vuln.Summary,
		})
	}

	return cves, nil
}

func (v *RegistryValidator) CheckLicense(license string) *LicenseCheck {
	license = strings.ToUpper(strings.TrimSpace(license))

	for _, l := range forbiddenLicenses {
		if strings.Contains(license, l) {
			return &LicenseCheck{
				License:  l,
				Allowed:  false,
				Category: "forbidden",
			}
		}
	}

	for _, l := range allowedLicenses {
		if strings.Contains(license, l) {
			return &LicenseCheck{
				License:  l,
				Allowed:  true,
				Category: "allowed",
			}
		}
	}

	return &LicenseCheck{
		License:  license,
		Allowed:  true,
		Category: "unknown",
	}
}

func (v *RegistryValidator) CheckTyposquatting(pkgName string) (bool, []string) {
	pkgLower := strings.ToLower(pkgName)

	for _, pattern := range typosquattingPatterns {
		for _, variant := range pattern.variants {
			distance := levenshteinDistance(pkgLower, strings.ToLower(variant))
			if distance <= 2 {
				return true, []string{pattern.original}
			}
		}

		if strings.HasPrefix(pkgLower, strings.ToLower(pattern.original)+"-") ||
			strings.HasSuffix(pkgLower, "-"+strings.ToLower(pattern.original)) {
			return true, []string{pattern.original}
		}
	}

	return false, []string{}
}

func (v *RegistryValidator) isSuspicious(pkgName string) bool {
	suspicious, _ := v.CheckTyposquatting(pkgName)
	return suspicious
}

func (v *RegistryValidator) calculateAge(timestamp string) int {
	if timestamp == "" {
		return 0
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}

	var t time.Time
	var err error

	for _, format := range formats {
		t, err = time.Parse(format, timestamp)
		if err == nil {
			break
		}
	}

	if err != nil {
		return 0
	}

	return int(time.Since(t).Hours() / 24)
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min(values ...int) int {
	min := math.MaxInt
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}
