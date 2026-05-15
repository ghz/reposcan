package tui

import (
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func getRepoIndex(repoIds []string, id string) int {
	for i, x := range repoIds {
		if x == id {
			return i
		}
	}
	return -1
}

func deleteRepo(repoIds []string, index int) []string {
	return append(repoIds[:index], repoIds[index+1:]...)
}

func shellEscapePath(path string) string {
	if runtime.GOOS == "windows" {
		return path
	}
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}

// applyViewMode updates the repos table to reflect the current view mode using
// the full (unfiltered) scan report stored in the model.
func (m *Model) applyViewMode() {
	switch m.viewMode {
	case ViewModeDirty:
		filtered := filterDirtyRepos(m.fullReport)
		m.reposTable.SetReport(filtered)
	case ViewModeAllRepos:
		m.reposTable.SetReport(m.fullReport)
	case ViewModeAllDirs:
		m.reposTable.SetFolders(m.fullReport.AllFolders, m.fullReport.RepoStates)
	case ViewModeNonRepoDirs:
		nonRepos := filterNonRepoFolders(m.fullReport.AllFolders)
		m.reposTable.SetFolders(nonRepos, nil)
	}
}

func makeAlert(t alerts.AlertType, message string) tea.Cmd {
	return func() tea.Msg {
		return alerts.AddAlertMsg{Msg: alerts.Alert{Type: t, Message: message}}
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// toggleFavorite adds or removes repoPath from the favorites list, then
// persists the updated config and refreshes the table.
func (m *Model) toggleFavorite(repoPath string) {
	favs := m.configs.Favorites
	found := false
	for i, p := range favs {
		if p == repoPath {
			favs = append(favs[:i], favs[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		favs = append(favs, repoPath)
	}
	m.configs.Favorites = favs

	if m.configs.ConfigFilePath != "" {
		_ = config.WriteToFile(m.configs, m.configs.ConfigFilePath)
	}

	m.reposTable.SetFavorites(favs)
}

func filterNonRepoFolders(folders []report.FolderEntry) []report.FolderEntry {
	result := make([]report.FolderEntry, 0)
	for _, f := range folders {
		if !f.IsRepo {
			result = append(result, f)
		}
	}
	return result
}

func filterDirtyRepos(r report.ScanReport) report.ScanReport {
	dirty := make([]report.RepoState, 0)
	for _, rs := range r.RepoStates {
		if rs.IsDirty() {
			dirty = append(dirty, rs)
		}
	}
	return report.ScanReport{
		Version:     r.Version,
		GeneratedAt: r.GeneratedAt,
		RepoStates:  dirty,
		AllFolders:  r.AllFolders,
		Warnings:    r.Warnings,
	}
}
