package repodetails

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.repoState == nil {
		return ""
	}

	if m.subMode == DetailsSubModeCommits {
		return m.viewCommits()
	}
	return m.viewFiles()
}

func (m *Model) viewFiles() string {
	style := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)

	lines := []string{
		fmt.Sprintf("%s %s", style.Render("Path:"), m.repoState.Path),
		style.Render("File Changes:"),
	}

	if len(m.repoState.UncommitedFiles) > 0 {
		files := m.repoState.UncommitedFiles

		maxToShow := m.height - len(lines) - 1
		trimmed := len(files) > maxToShow

		if trimmed {
			files = files[:maxToShow]
		}

		for _, f := range files {
			lines = append(lines, "  "+m.theme.Styles.Muted.Render(f))
		}

		if trimmed {
			more := len(m.repoState.UncommitedFiles) - maxToShow
			lines = append(lines, m.theme.Styles.Muted.Render("  ... (+"+strconv.Itoa(more)+" more)"))
		}
	} else {
		lines = append(lines, m.theme.Styles.Muted.Render("    no changes"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *Model) viewCommits() string {
	style := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)

	lines := []string{
		fmt.Sprintf("%s %s", style.Render("Path:"), m.repoState.Path),
		style.Render("Recent Commits:"),
	}

	if len(m.commits) == 0 {
		lines = append(lines, m.theme.Styles.Muted.Render("    no commits"))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	commits := m.commits
	maxToShow := m.height - len(lines) - 1
	trimmed := len(commits) > maxToShow
	if trimmed {
		commits = commits[:maxToShow]
	}

	for _, c := range commits {
		lines = append(lines, "  "+m.theme.Styles.Muted.Render(c))
	}

	if trimmed {
		more := len(m.commits) - maxToShow
		lines = append(lines, m.theme.Styles.Muted.Render("  ... (+"+strconv.Itoa(more)+" more)"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
