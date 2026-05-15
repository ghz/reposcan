package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) generateCreateRepoPopup() string {
	t := m.theme

	keyStyle := t.Styles.Base.Bold(true).Foreground(t.Colors.Accent)
	textStyle := t.Styles.PopupText
	mutedStyle := t.Styles.PopupText.Foreground(t.Colors.Muted)

	title := t.Styles.PopupHeader.Render("Nouveau repo")

	nameRow := lipgloss.JoinHorizontal(lipgloss.Left,
		textStyle.Render("Nom: "),
		m.createRepoNameInput.View(),
	)

	separator := mutedStyle.Render("──────────────────────────────────")

	actions := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[l]"), textStyle.Render("  Local uniquement")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[p]"), textStyle.Render("  GitHub — privé")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("[u]"), textStyle.Render("  GitHub — public")),
	)

	cancel := mutedStyle.Render("[esc] Annuler")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		nameRow,
		"",
		separator,
		"",
		actions,
		"",
		cancel,
	)

	return t.Styles.Popup.Render(content)
}
