package repodetails

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

func (m *Model) View() string {
	if m.repoState == nil {
		return ""
	}

	switch m.subMode {
	case DetailsSubModeCommits:
		return m.viewCommits()
	case DetailsSubModeReadme:
		return m.viewReadme()
	default:
		return m.viewFiles()
	}
}

// join truncates every line to the panel width (ANSI-aware, with an ellipsis
// tail) before stacking them, so long commit messages or README lines can't
// wrap and push the rest of the layout off-screen.
func (m *Model) join(lines []string) string {
	if m.width > 0 {
		for i, l := range lines {
			lines[i] = truncate.StringWithTail(l, uint(m.width), "…")
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// viewTabs renders the file-changes / recent-commits / readme switch as a tab
// bar, with the active tab highlighted and a hint on the same line.
func (m *Model) viewTabs() string {
	active := m.theme.Styles.Base.
		Foreground(m.theme.Colors.Accent).
		Background(m.theme.Colors.TableAltRow).
		Bold(true).
		Padding(0, 1)
	inactive := m.theme.Styles.Muted.Padding(0, 1)

	tab := func(mode DetailsSubMode, label string) string {
		if m.subMode == mode {
			return active.Render(label)
		}
		return inactive.Render(label)
	}

	hint := m.theme.Styles.Muted.Render("  ← → to switch")
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		tab(DetailsSubModeFiles, "File changes"),
		tab(DetailsSubModeCommits, "Recent commits"),
		tab(DetailsSubModeReadme, "README"),
		hint,
	)
}

func (m *Model) viewFiles() string {
	style := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)

	lines := []string{
		fmt.Sprintf("%s %s", style.Render("Path:"), m.repoState.Path),
		m.viewTabs(),
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

	return m.join(lines)
}

func (m *Model) viewCommits() string {
	style := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)

	lines := []string{
		fmt.Sprintf("%s %s", style.Render("Path:"), m.repoState.Path),
		m.viewTabs(),
	}

	if len(m.commits) == 0 {
		lines = append(lines, m.theme.Styles.Muted.Render("    no commits"))
		return m.join(lines)
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

	return m.join(lines)
}

func (m *Model) viewReadme() string {
	style := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)

	lines := []string{
		fmt.Sprintf("%s %s", style.Render("Path:"), m.repoState.Path),
		m.viewTabs(),
	}

	if len(m.readme) == 0 {
		lines = append(lines, m.theme.Styles.Muted.Render("    no README found"))
		return m.join(lines)
	}

	content := m.readme
	maxToShow := m.height - len(lines) - 1
	trimmed := len(content) > maxToShow
	if trimmed {
		content = content[:maxToShow]
	}

	for _, l := range content {
		lines = append(lines, "  "+m.theme.Styles.Muted.Render(l))
	}

	if trimmed {
		more := len(m.readme) - maxToShow
		lines = append(lines, m.theme.Styles.Muted.Render("  ... (+"+strconv.Itoa(more)+" more)"))
	}

	return m.join(lines)
}
