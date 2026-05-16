package tui

import "github.com/charmbracelet/lipgloss"

func (m *Model) generateGitMenuPopup() string {
	t := m.theme

	keyStyle := t.Styles.Base.Bold(true).Foreground(t.Colors.Accent)
	textStyle := t.Styles.PopupText
	mutedStyle := t.Styles.PopupText.Foreground(t.Colors.Muted)
	valueStyle := t.Styles.PopupText.Bold(true).Foreground(t.Colors.Info)

	title := t.Styles.PopupHeader.Render("Git")
	separator := mutedStyle.Render("──────────────────────────────────")

	// Make the target explicit: these actions always run on the repo's
	// checked-out branch, even when a different branch row is highlighted.
	project, branch := "-", "-"
	if rs := m.reposTable.GetCurrentRepoState(); rs != nil {
		project = rs.Repo
		branch = rs.Branch
	}
	target := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, mutedStyle.Render("Project  "), valueStyle.Render(project)),
		lipgloss.JoinHorizontal(lipgloss.Left, mutedStyle.Render("Branch   "), valueStyle.Render(branch)),
	)

	actions := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[1]"), textStyle.Render("  Quick save")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[2]"), textStyle.Render("  Commit…")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[3]"), textStyle.Render("  Push")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[4]"), textStyle.Render("  Pull")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[5]"), textStyle.Render("  Fetch")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[6]"), textStyle.Render("  Open remote")),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		target,
		"",
		separator,
		"",
		textStyle.Render("Action:"),
		"",
		actions,
		"",
		separator,
		"",
		mutedStyle.Render("[esc] Cancel"),
	)
	return t.Styles.Popup.Render(content)
}
