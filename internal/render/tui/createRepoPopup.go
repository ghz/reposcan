package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) generateCreateRepoPopup() string {
	t := m.theme

	keyStyle := t.Styles.Base.Bold(true).Foreground(t.Colors.Accent)
	textStyle := t.Styles.PopupText
	mutedStyle := t.Styles.PopupText.Foreground(t.Colors.Muted)

	title := t.Styles.PopupHeader.Render("New repo")
	separator := mutedStyle.Render("──────────────────────────────────")

	if m.createStep == stepChooseKind {
		actions := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[1]"), textStyle.Render("  Local only")),
			lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[2]"), textStyle.Render("  GitHub private")),
			lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[3]"), textStyle.Render("  GitHub public")),
		)

		content := lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			textStyle.Render("Repo type:"),
			"",
			actions,
			"",
			separator,
			"",
			mutedStyle.Render("[esc] Cancel"),
		)
		return t.Styles.Popup.Render(content)
	}

	nameRow := lipgloss.JoinHorizontal(lipgloss.Left,
		textStyle.Render("Name: "),
		m.createRepoNameInput.View(),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		mutedStyle.Render("Type: "+m.createKind.Label()),
		"",
		nameRow,
		"",
		separator,
		"",
		mutedStyle.Render("[enter] Create   [esc] Back"),
	)
	return t.Styles.Popup.Render(content)
}
