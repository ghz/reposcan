package repodetails

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

func (m *Model) View() string {
	if m.repoState == nil {
		return ""
	}

	switch m.subMode {
	case DetailsSubModeDiff:
		return m.viewDiff()
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
		tab(DetailsSubModeDiff, "Diff"),
		tab(DetailsSubModeCommits, "Recent commits"),
		tab(DetailsSubModeReadme, "README"),
		hint,
	)
}

func (m *Model) viewFiles() string {
	content := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)

	body := make([]string, 0, len(m.repoState.UncommitedFiles))
	for _, f := range m.repoState.UncommitedFiles {
		body = append(body, "  "+content.Render(f))
	}
	return m.render(body, "    no changes")
}

// viewDiff renders the colored `git diff` output. Lines already carry git's
// own ANSI color codes, so they are emitted as-is (only indented) rather than
// re-wrapped in a lipgloss style, which would clobber the embedded colors.
func (m *Model) viewDiff() string {
	body := make([]string, 0, len(m.diff))
	for _, l := range m.diff {
		body = append(body, "  "+l)
	}
	return m.render(body, "    no changes")
}

func (m *Model) viewCommits() string {
	content := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)

	body := make([]string, 0, len(m.commits))
	for _, c := range m.commits {
		body = append(body, "  "+content.Render(c))
	}
	return m.render(body, "    no commits")
}

func (m *Model) viewReadme() string {
	content := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)

	body := make([]string, 0, len(m.readme))
	for _, l := range m.readme {
		body = append(body, "  "+content.Render(l))
	}
	return m.render(body, "    no README found")
}

// render assembles the details panel: the "Path:" line and tab bar, then the
// slice of body lines visible at the current scroll offset. When the body is
// taller than the panel, the last row shows how many lines are hidden above
// and below and how to scroll.
func (m *Model) render(body []string, emptyMsg string) string {
	infoStyle := m.theme.Styles.Base.Foreground(m.theme.Colors.Info)
	header := []string{
		fmt.Sprintf("%s %s", infoStyle.Render("Path:"), m.repoState.Path),
		m.viewTabs(),
	}

	if len(body) == 0 {
		return m.join(append(header, m.theme.Styles.Muted.Render(emptyMsg)))
	}

	avail := m.height - len(header) - 1
	if avail < 1 {
		avail = 1
	}

	// Everything fits: no scrolling, no hint line.
	if len(body) <= avail {
		return m.join(append(header, body...))
	}

	// Clamp the offset; the panel may have grown since the last scroll.
	maxOffset := len(body) - avail
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	window := body[m.scrollOffset : m.scrollOffset+avail]
	above := m.scrollOffset
	below := len(body) - (m.scrollOffset + avail)
	hint := fmt.Sprintf("  ↑ %d above · ↓ %d below — PgUp/PgDn", above, below)

	lines := append(header, window...)
	lines = append(lines, m.theme.Styles.Muted.Render(hint))
	return m.join(lines)
}
