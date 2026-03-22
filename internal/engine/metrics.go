package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/TU_ORG/fenrir/internal/graph"
)

type Metrics struct {
	graph      *graph.Graph
	dataDir    string
	mu         sync.RWMutex
	adoption   *AdoptionMetrics
	quality    *QualityMetrics
	impact     *ImpactMetrics
	startTime  time.Time
}

type AdoptionMetrics struct {
	Installations      int       `json:"installations"`
	ActiveProjects     int       `json:"active_projects"`
	GitHubStars        int       `json:"github_stars"`
	LastInstallation   time.Time `json:"last_installation"`
	Installations3M    int       `json:"installations_3_months"`
	Installations6M    int       `json:"installations_6_months"`
	Projects3M         int       `json:"projects_3_months"`
	Projects6M         int       `json:"projects_6_months"`
	Stars3M            int       `json:"stars_3_months"`
	Stars6M            int       `json:"stars_6_months"`
}

type QualityMetrics struct {
	InitTime           float64 `json:"init_time_ms"`
	MCPResponseP99     float64 `json:"mcp_response_p99_ms"`
	CrashRate          float64 `json:"crash_rate_percent"`
	TyposquattingPrecision float64 `json:"typosquatting_precision_percent"`
	InjectGuardFP      float64 `json:"inject_guard_false_positive_percent"`
	MemFindResponse    float64 `json:"mem_find_response_ms"`
	PkgCheckCacheHit   float64 `json:"pkg_check_cache_hit_ms"`
	PkgCheckCacheMiss  float64 `json:"pkg_check_cache_miss_ms"`
	DriftCalcTime      float64 `json:"drift_calc_ms"`
	BinarySizeMB       float64 `json:"binary_size_mb"`
	StartupTime        float64 `json:"startup_time_ms"`
}

type ImpactMetrics struct {
	CVEPackagesBlocked int     `json:"cve_packages_blocked"`
	CVEPackagesAllowed int     `json:"cve_packages_allowed"`
	DriftReduction     float64 `json:"drift_reduction_percent"`
	SessionEndRate     float64 `json:"session_end_completion_rate"`
	Week1DriftAvg      float64 `json:"week1_drift_average"`
	Week8DriftAvg      float64 `json:"week8_drift_average"`
}

type PerformanceSnapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	MemFindP50     float64   `json:"mem_find_p50_ms"`
	MemFindP99     float64   `json:"mem_find_p99_ms"`
	PkgCheckP50    float64   `json:"pkg_check_p50_ms"`
	PkgCheckP99    float64   `json:"pkg_check_p99_ms"`
	TotalRequests  int64     `json:"total_requests"`
	Errors         int64     `json:"errors"`
	SuccessRate    float64   `json:"success_rate_percent"`
}

type SessionMetrics struct {
	SessionID         string    `json:"session_id"`
	Goal              string    `json:"goal"`
	StartedAt         time.Time `json:"started_at"`
	EndedAt           time.Time `json:"ended_at"`
	DurationMs        int64     `json:"duration_ms"`
	MemSaveCalls      int       `json:"mem_save_calls"`
	PkgCheckCalls     int       `json:"pkg_check_calls"`
	ArchVerifyCalls   int       `json:"arch_verify_calls"`
	DriftScore        float64   `json:"drift_score"`
	ArchViolations    int       `json:"arch_violations"`
	Warnings          int       `json:"warnings"`
	Errors            int       `json:"errors"`
	Completed         bool      `json:"completed"`
}

func New(dataDir string) (*Metrics, error) {
	m := &Metrics{
		dataDir:   dataDir,
		startTime: time.Now(),
		adoption:  &AdoptionMetrics{},
		quality:   &QualityMetrics{},
		impact:    &ImpactMetrics{},
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	m.loadMetrics()

	return m, nil
}

func (m *Metrics) loadMetrics() {
	metricsFile := filepath.Join(m.dataDir, "metrics.json")
	
	data, err := os.ReadFile(metricsFile)
	if err != nil {
		return
	}

	var saved struct {
		Adoption *AdoptionMetrics `json:"adoption"`
		Quality  *QualityMetrics  `json:"quality"`
		Impact   *ImpactMetrics   `json:"impact"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}

	if saved.Adoption != nil {
		m.adoption = saved.Adoption
	}
	if saved.Quality != nil {
		m.quality = saved.Quality
	}
	if saved.Impact != nil {
		m.impact = saved.Impact
	}
}

func (m *Metrics) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	metricsFile := filepath.Join(m.dataDir, "metrics.json")

	data, err := json.MarshalIndent(struct {
		Adoption *AdoptionMetrics `json:"adoption"`
		Quality  *QualityMetrics  `json:"quality"`
		Impact   *ImpactMetrics   `json:"impact"`
	}{
		Adoption: m.adoption,
		Quality:  m.quality,
		Impact:   m.impact,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metricsFile, data, 0644)
}

func (m *Metrics) RecordInstallation() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.adoption.Installations++
	m.adoption.LastInstallation = time.Now()
	m.adoption.Installations3M++
	m.adoption.Installations6M++

	m.Save()
}

func (m *Metrics) RecordActiveProject(projectID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.adoption.ActiveProjects++
	m.adoption.Projects3M++
	m.adoption.Projects6M++

	m.Save()
}

func (m *Metrics) RecordInitTime(durationMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.quality.InitTime == 0 || durationMs < m.quality.InitTime {
		m.quality.InitTime = durationMs
	}

	m.Save()
}

func (m *Metrics) RecordStartupTime(durationMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.quality.StartupTime = durationMs

	if durationMs < 100 {
		m.checkQualityThreshold("startup_time", durationMs, 100)
	}

	m.Save()
}

func (m *Metrics) RecordMemFindTime(durationMs float64, cached bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cached {
		if m.quality.MemFindResponse == 0 || durationMs < m.quality.MemFindResponse {
			m.quality.MemFindResponse = durationMs
		}
	}

	m.Save()
}

func (m *Metrics) RecordPkgCheckTime(durationMs float64, cacheHit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cacheHit {
		m.quality.PkgCheckCacheHit = durationMs
	} else {
		m.quality.PkgCheckCacheMiss = durationMs
	}

	m.Save()
}

func (m *Metrics) RecordCrash() {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessions := m.getTotalSessions()
	if sessions > 0 {
		m.quality.CrashRate = (1.0 / float64(sessions)) * 100
	}

	m.Save()
}

func (m *Metrics) RecordBinarySize(sizeMB float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.quality.BinarySizeMB = sizeMB

	m.Save()
}

func (m *Metrics) RecordCVEPackage(blocked bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if blocked {
		m.impact.CVEPackagesBlocked++
	} else {
		m.impact.CVEPackagesAllowed++
	}

	m.Save()
}

func (m *Metrics) RecordSessionCompleted(completed bool, driftScore float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if completed {
		m.impact.SessionEndRate = 1.0
	}

	if driftScore > 0 {
		if m.impact.Week1DriftAvg == 0 {
			m.impact.Week1DriftAvg = driftScore
		} else {
			m.impact.Week1DriftAvg = (m.impact.Week1DriftAvg + driftScore) / 2
		}
	}

	m.Save()
}

func (m *Metrics) getTotalSessions() int {
	return 0
}

func (m *Metrics) checkQualityThreshold(metric string, value, threshold float64) {
	if value > threshold {
		m.quality.CrashRate = value
	}
}

func (m *Metrics) GetAdoption() *AdoptionMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.adoption
}

func (m *Metrics) GetQuality() *QualityMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.quality
}

func (m *Metrics) GetImpact() *ImpactMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.impact
}

func (m *Metrics) GetAll() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"adoption": m.adoption,
		"quality":  m.quality,
		"impact":   m.impact,
		"uptime":   time.Since(m.startTime).String(),
	}
}

func (m *Metrics) CheckTargets() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	checks := map[string]bool{
		"init_under_60s":          m.quality.InitTime > 0 && m.quality.InitTime < 60000,
		"startup_under_100ms":     m.quality.StartupTime > 0 && m.quality.StartupTime < 100,
		"mcp_p99_under_200ms":     m.quality.MCPResponseP99 > 0 && m.quality.MCPResponseP99 < 200,
		"crash_rate_under_0.1":   m.quality.CrashRate < 0.1,
		"typosquatting_over_95":   m.quality.TyposquattingPrecision > 95,
		"inject_fp_under_5":       m.quality.InjectGuardFP < 5,
		"binary_under_20mb":       m.quality.BinarySizeMB > 0 && m.quality.BinarySizeMB < 20,
		"installations_3m_500":    m.adoption.Installations3M >= 500,
		"installations_6m_2000":   m.adoption.Installations6M >= 2000,
		"projects_3m_100":         m.adoption.Projects3M >= 100,
		"projects_6m_500":         m.adoption.Projects6M >= 500,
	}

	return checks
}

func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

func (m *Metrics) ExportJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return json.MarshalIndent(struct {
		Adoption  *AdoptionMetrics  `json:"adoption"`
		Quality   *QualityMetrics   `json:"quality"`
		Impact    *ImpactMetrics    `json:"impact"`
		Uptime    string            `json:"uptime"`
		Timestamp time.Time         `json:"timestamp"`
	}{
		Adoption:  m.adoption,
		Quality:   m.quality,
		Impact:    m.impact,
		Uptime:    time.Since(m.startTime).String(),
		Timestamp: time.Now(),
	}, "", "  ")
}
