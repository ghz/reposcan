package tui

import "github.com/charmbracelet/lipgloss"

func (m *Model) generateGitMenuPopup() string {
	t := m.theme

	keyStyle := t.Styles.Base.Bold(true).Foreground(t.Colors.Accent)
	textStyle := t.Styles.PopupText
	mutedStyle := t.Styles.PopupText.Foreground(t.Colors.Muted)

	title := t.Styles.PopupHeader.Render("Git")
	separator := mutedStyle.Render("──────────────────────────────────")

	actions := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[1]"), textStyle.Render("  Quick save")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[2]"), textStyle.Render("  Push")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[3]"), textStyle.Render("  Pull")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[4]"), textStyle.Render("  Fetch")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[5]"), textStyle.Render("  Open remote")),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
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
