package tui

import (
	"runtime"
	"strings"

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
	m.reposTable.SetTitle(m.viewMode.Label())
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
