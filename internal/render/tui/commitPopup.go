package tui

import "github.com/charmbracelet/lipgloss"

func (m *Model) generateCommitPopup() string {
	t := m.theme

	textStyle := t.Styles.PopupText
	mutedStyle := t.Styles.PopupText.Foreground(t.Colors.Muted)
	valueStyle := t.Styles.PopupText.Bold(true).Foreground(t.Colors.Info)

	title := t.Styles.PopupHeader.Render("Commit")
	separator := mutedStyle.Render("──────────────────────────────────")

	project := "-"
	if rs := m.reposTable.GetCurrentRepoState(); rs != nil {
		project = rs.Repo
	}

	messageRow := lipgloss.JoinHorizontal(lipgloss.Left,
		textStyle.Render("Message: "),
		m.commitMessageInput.View(),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, mutedStyle.Render("Project  "), valueStyle.Render(project)),
		"",
		separator,
		"",
		mutedStyle.Render("Stages all changes, then commits. No push."),
		"",
		messageRow,
		"",
		separator,
		"",
		mutedStyle.Render("[enter] Commit (empty = wip)   [esc] Back"),
	)
	return t.Styles.Popup.Render(content)
}
