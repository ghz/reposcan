package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func (m Model) generateDeleteRepoPopup() string {
	targetName, targetPath, rs := m.deleteTarget()

	t := m.theme
	titleStyle := t.Styles.Base.Bold(true).Foreground(t.Colors.Accent)
	textStyle := t.Styles.Base.Foreground(t.Colors.Foreground)
	mutedStyle := t.Styles.Muted
	inputStyle := t.Styles.Base.Foreground(t.Colors.Foreground)
	statusStyle := t.Styles.Base.Foreground(t.Colors.Warning)
	if rs != nil && !rs.IsDirty() {
		statusStyle = t.Styles.Base.Foreground(t.Colors.Success)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Delete local folder?"),
		"",
		textStyle.Render(targetName),
		mutedStyle.Render(targetPath),
		"",
		statusStyle.Render(deleteRepoStatusText(rs)),
		"",
		mutedStyle.Render("This only deletes the local folder. It does not delete any remote/GitHub repository."),
		mutedStyle.Render("Type YES and press Enter to permanently delete this folder."),
		inputStyle.Render(m.deleteConfirmInput.View()),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Colors.Accent).
		Padding(1, 2).
		Width(64).
		Render(content)
}

func (m Model) deleteTarget() (string, string, *report.RepoState) {
	path := m.reposTable.GetCurrentPath()
	name := filepath.Base(path)

	if entry := m.reposTable.GetCurrentFolderEntry(); entry != nil {
		if entry.Name != "" {
			name = entry.Name
		}
	}

	rs := m.reposTable.GetCurrentRepoState()
	if rs != nil {
		if rs.Repo != "" {
			name = rs.Repo
		}
		if rs.Path != "" {
			path = rs.Path
		}
	}

	return name, path, rs
}

func deleteRepoStatusText(rs *report.RepoState) string {
	if rs == nil {
		return "This folder is not a Git repository."
	}

	lines := make([]string, 0, 3)
	if len(rs.UncommitedFiles) > 0 {
		lines = append(lines, fmt.Sprintf("%d uncommitted file(s)", len(rs.UncommitedFiles)))
	}

	ahead, behind := totalAheadBehind(rs.RemoteStatus)
	if ahead > 0 {
		lines = append(lines, fmt.Sprintf("%d commit(s) to push", ahead))
	}
	if behind > 0 {
		lines = append(lines, fmt.Sprintf("%d commit(s) to pull", behind))
	}

	if len(lines) == 0 {
		return "No local changes, commits to push, or commits to pull detected."
	}
	return "Warning: " + strings.Join(lines, " | ")
}

func totalAheadBehind(remotes []report.RemoteStatus) (int, int) {
	ahead := 0
	behind := 0
	for _, remote := range remotes {
		ahead += remote.Ahead
		behind += remote.Behind
	}
	return ahead, behind
}
