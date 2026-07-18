package agentd

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agent"
)

// dashboardHandler renders the SLO monitoring dashboard.
func (s *Server) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats := s.aggregateStats()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(renderDashboard(stats)))
}

// DashboardStats holds aggregated SLO statistics.
type DashboardStats struct {
	TotalRuns      int            `json:"total_runs"`
	SuccessRate    float64        `json:"success_rate"`
	AvgDurationMs  int64          `json:"avg_duration_ms"`
	ActiveRuns     int            `json:"active_runs"`
	QueuedRuns     int            `json:"queued_runs"`
	BySkill        map[string]int `json:"by_skill"`
	RecentRuns     []RunSummary   `json:"recent_runs"`
}

// RunSummary is a summary of a single run.
type RunSummary struct {
	RunID     string `json:"run_id"`
	Status    string `json:"status"`
	Product   string `json:"product"`
	Step      string `json:"step"`
	CreatedAt string `json:"created_at"`
}

// aggregateStats scans all runs and computes statistics.
func (s *Server) aggregateStats() *DashboardStats {
	stats := &DashboardStats{
		BySkill: make(map[string]int),
	}

	runDir := filepath.Join(s.root, ".runtime", "agent", "runs")
	entries, err := os.ReadDir(runDir)
	if err != nil {
		return stats
	}

	var totalDuration int64
	var successCount int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runID := entry.Name()
		state, err := agent.LoadState(s.root, runID)
		if err != nil || state == nil {
			continue
		}

		stats.TotalRuns++
		runStatus := determineStatus(state)

		// Count by skill
		if state.Triage != nil {
			stats.BySkill[state.Triage.PrimarySkill]++
		}

		// Count active/queued
		if runStatus == "running" {
			stats.ActiveRuns++
		} else if runStatus == "queued" {
			stats.QueuedRuns++
		}

		// Success rate
		if runStatus == "completed" {
			successCount++
		}

		// Duration (approximate from run ID timestamp)
		ts, _ := parseRunTimestamp(runID)
		if !ts.IsZero() {
			totalDuration += time.Since(ts).Milliseconds()
		}

		// Recent runs (last 10)
		if len(stats.RecentRuns) < 10 {
			stats.RecentRuns = append(stats.RecentRuns, RunSummary{
				RunID:     runID,
				Status:    runStatus,
				Product:   state.Payload.ProductHint,
				Step:      state.CurrentStep.String(),
				CreatedAt: ts.UTC().Format(time.RFC3339),
			})
		}
	}

	if stats.TotalRuns > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalRuns)
		stats.AvgDurationMs = totalDuration / int64(stats.TotalRuns)
	}

	return stats
}

func parseRunTimestamp(runID string) (time.Time, error) {
	var ts int64
	for _, c := range runID {
		if c >= '0' && c <= '9' {
			ts = ts*10 + int64(c-'0')
		}
	}
	if ts == 0 {
		return time.Time{}, nil
	}
	return time.Unix(0, ts), nil
}

func renderDashboard(stats *DashboardStats) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Agent Dashboard</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 20px; background: #f5f5f5; }
        .card { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; margin-bottom: 10px; }
        h2 { color: #666; font-size: 14px; text-transform: uppercase; margin-bottom: 15px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; }
        .stat { text-align: center; }
        .stat-value { font-size: 32px; font-weight: bold; color: #007bff; }
        .stat-label { font-size: 12px; color: #666; margin-top: 5px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f8f9fa; font-weight: 600; }
        .status { padding: 4px 8px; border-radius: 4px; font-size: 12px; }
        .status-running { background: #28a745; color: white; }
        .status-completed { background: #6c757d; color: white; }
        .status-failed { background: #dc3545; color: white; }
        .status-paused { background: #ffc107; color: #333; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Agent Dashboard</h1>
        <p style="color: #666;">SLO Monitoring Panel</p>
    </div>

    <div class="card">
        <h2>SLO Overview</h2>
        <div class="stats">
            <div class="stat">
                <div class="stat-value">` + formatInt(stats.TotalRuns) + `</div>
                <div class="stat-label">Total Runs</div>
            </div>
            <div class="stat">
                <div class="stat-value">` + formatPercent(stats.SuccessRate) + `</div>
                <div class="stat-label">Success Rate</div>
            </div>
            <div class="stat">
                <div class="stat-value">` + formatDuration(stats.AvgDurationMs) + `</div>
                <div class="stat-label">Avg Duration</div>
            </div>
            <div class="stat">
                <div class="stat-value">` + formatInt(stats.ActiveRuns) + `</div>
                <div class="stat-label">Active Runs</div>
            </div>
        </div>
    </div>

    <div class="card">
        <h2>Recent Runs</h2>
        <table>
            <thead>
                <tr>
                    <th>Run ID</th>
                    <th>Status</th>
                    <th>Product</th>
                    <th>Step</th>
                    <th>Created</th>
                </tr>
            </thead>
            <tbody>
                ` + renderRunsTable(stats.RecentRuns) + `
            </tbody>
        </table>
    </div>

    <div class="card">
        <h2>Runs by Skill</h2>
        <table>
            <thead>
                <tr>
                    <th>Skill</th>
                    <th>Count</th>
                </tr>
            </thead>
            <tbody>
                ` + renderSkillTable(stats.BySkill) + `
            </tbody>
        </table>
    </div>
</body>
</html>`
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func formatPercent(f float64) string {
	return formatInt(int(f*100)) + "%"
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return formatInt(int(ms)) + "ms"
	}
	return formatInt(int(ms/1000)) + "s"
}

func runsTable(runs []RunSummary) string {
	s := ""
	for _, r := range runs {
		s += `<tr>
            <td>` + r.RunID + `</td>
            <td><span class="status status-` + r.Status + `">` + r.Status + `</span></td>
            <td>` + r.Product + `</td>
            <td>` + r.Step + `</td>
            <td>` + r.CreatedAt + `</td>
        </tr>`
	}
	return s
}

func skillTable(skills map[string]int) string {
	s := ""
	for skill, count := range skills {
		s += `<tr><td>` + skill + `</td><td>` + formatInt(count) + `</td></tr>`
	}
	return s
}

func renderRunsTable(runs []RunSummary) string {
	return runsTable(runs)
}

func renderSkillTable(skills map[string]int) string {
	return skillTable(skills)
}
