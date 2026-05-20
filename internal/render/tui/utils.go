package tui

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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

func (m *Model) removePathFromReport(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}

	repoStates := make([]report.RepoState, 0, len(m.fullReport.RepoStates))
	for _, rs := range m.fullReport.RepoStates {
		if !samePath(rs.Path, path) {
			repoStates = append(repoStates, rs)
		}
	}
	m.fullReport.RepoStates = repoStates

	folders := make([]report.FolderEntry, 0, len(m.fullReport.AllFolders))
	for _, f := range m.fullReport.AllFolders {
		if !samePath(f.Path, path) {
			folders = append(folders, f)
		}
	}
	m.fullReport.AllFolders = folders

	favorites := make([]string, 0, len(m.configs.Favorites))
	for _, favorite := range m.configs.Favorites {
		if !samePath(favorite, path) {
			favorites = append(favorites, favorite)
		}
	}
	m.configs.Favorites = favorites
	m.reposTable.SetFavorites(favorites)
	m.applyViewMode()
}

func samePath(a string, b string) bool {
	left := filepath.Clean(a)
	right := filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func makeAlert(t alerts.AlertType, message string) tea.Cmd {
	return func() tea.Msg {
		return alerts.AddAlertMsg{Msg: alerts.Alert{Type: t, Message: message}}
	}
}

func rollbackCreatedGitDirOnError(repoPath string, fn func() error) error {
	hadGitDir, err := gitDirExists(repoPath)
	if err != nil {
		return err
	}

	err = fn()
	if err == nil || hadGitDir {
		return err
	}

	gitDir := filepath.Join(repoPath, ".git")
	if cleanupErr := os.RemoveAll(gitDir); cleanupErr != nil {
		return errors.Join(err, cleanupErr)
	}
	return err
}

func gitDirExists(repoPath string) (bool, error) {
	_, err := os.Lstat(filepath.Join(repoPath, ".git"))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
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

func commandForOpenPath(name string, path string) *exec.Cmd {
	return exec.Command(name, path)
}

func openFileManager(path string) error {
	cmd := commandForFileManager(runtime.GOOS, path)
	return cmd.Start()
}

func commandForFileManager(goos string, path string) *exec.Cmd {
	switch goos {
	case "windows":
		return exec.Command("explorer", path)
	case "darwin":
		return exec.Command("open", path)
	default:
		return exec.Command("xdg-open", path)
	}
}

// openTerminal launches a terminal emulator with its working directory set to
// path. When terminal is non-empty it is used as the command (the path is
// passed as its sole argument); otherwise a platform default is used.
func openTerminal(terminal string, path string) error {
	cmd := commandForTerminal(runtime.GOOS, terminal, path)
	err := cmd.Start()
	if err != nil && strings.TrimSpace(terminal) == "" && runtime.GOOS == "windows" {
		// Windows Terminal (wt) may not be installed; fall back to cmd.
		fallback := exec.Command("cmd", "/c", "start", "cmd", "/k", "cd", "/d", path)
		return fallback.Start()
	}
	return err
}

func commandForTerminal(goos string, terminal string, path string) *exec.Cmd {
	if strings.TrimSpace(terminal) != "" {
		return exec.Command(terminal, path)
	}
	switch goos {
	case "windows":
		// Windows Terminal opens directly at the given directory.
		return exec.Command("wt", "-d", path)
	case "darwin":
		return exec.Command("open", "-a", "Terminal", path)
	default:
		return exec.Command("x-terminal-emulator", "--working-directory="+path)
	}
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
