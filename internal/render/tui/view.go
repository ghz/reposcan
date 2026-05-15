package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/render/tui/overlay"
)

func (m Model) View() string {
	if m.loading {
		return "Loading..."
	}

	footer := m.getFooterView()

	// Reserve a fixed footer height based on the longest possible keybinding
	// line so the body never shifts when contextual keybindings (like "n")
	// appear or disappear.
	footerHeight := m.stableFooterHeight(footer)
	footer = lipgloss.NewStyle().Width(m.width).Height(footerHeight).Render(footer)

	bodyHeight := m.height - footerHeight

	reposTableHeight := bodyHeight * sizeReposTableHeightPercent / 100
	m.reposTable = m.reposTable.UpdateWindowSize(m.width, reposTableHeight)
	reposTable := m.reposTable.View()

	m.repoDetails.UpdateData(m.reposTable.GetCurrentRepoState())
	m.repoDetails.UpdateSize(bodyHeight - reposTableHeight)
	reposDetails := m.repoDetails.View()

	body := lipgloss.JoinVertical(lipgloss.Left, reposTable, reposDetails)
	body = lipgloss.NewStyle().
		Height(bodyHeight).
		MaxHeight(bodyHeight).
		Render(body)

	view := lipgloss.JoinVertical(lipgloss.Left, body, footer)

	view = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(view)

	view = m.renderAlerts(view, m.alerts.AlertStates(m.width, m.height))

	if m.currentFocus() == FocusHelpPopup {
		helpView := generateHelpPopup(m.theme, reposTableKeybindings)
		view = overlay.PlaceOverlayWithPosition(
			overlay.OverlayPositionCenter,
			m.width, m.height,
			helpView, view,
			true,
			overlay.WithWhitespaceChars(" "),
		)
	}

	if m.currentFocus() == FocusCreateRepoPopup {
		popup := m.generateCreateRepoPopup()
		view = overlay.PlaceOverlayWithPosition(
			overlay.OverlayPositionCenter,
			m.width, m.height,
			popup, view,
			true,
			overlay.WithWhitespaceChars(" "),
		)
	}

	return view
}

func (m *Model) getFooterView() string {
	if m.IsReposFilterVisible() {
		//if m.reposFilter.show {
		return m.theme.Styles.Base.
			Foreground(m.theme.Colors.Foreground).
			Render(m.reposFilter.View())
	}

	return m.generateKeybindingsFooterView()
}

func (m *Model) generateKeybindingsFooterView() string {
	keybindings := m.keybindings()
	if m.currentFocus() == FocusReposTable {
		keybindings = append(keybindings, common.Keybinding{Key: "?", ShortDesc: "Help"})
	}
	return m.renderKeybindingsLine(keybindings)
}

func (m *Model) renderKeybindingsLine(keybindings []common.Keybinding) string {
	kbStyle := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)
	mutedStyle := m.theme.Styles.Muted

	var sb strings.Builder
	for i, kb := range keybindings {
		sb.WriteString(mutedStyle.Render(kb.ShortDesc))
		sb.WriteString(mutedStyle.Render(": "))
		sb.WriteString(kbStyle.Render(kb.Key))
		if i < len(keybindings)-1 {
			sb.WriteString(mutedStyle.Render(" | "))
		}
	}
	return m.theme.Styles.Muted.Render(sb.String())
}

// stableFooterHeight returns the height to reserve for the footer. It measures
// the line height *after* wrapping to the terminal width, and uses the longest
// possible keybinding line (with every contextual binding present) so the body
// never shifts when contextual keybindings appear or disappear.
func (m *Model) stableFooterHeight(actualFooter string) int {
	width := m.width
	if width <= 0 {
		width = 80
	}
	measure := func(s string) int {
		return lipgloss.Height(lipgloss.NewStyle().Width(width).Render(s))
	}

	actual := measure(actualFooter)
	if m.IsReposFilterVisible() || m.currentFocus() != FocusReposTable {
		return actual
	}

	fullKbs := make([]common.Keybinding, len(reposTableKeybindings)+1)
	copy(fullKbs, reposTableKeybindings)
	fullKbs[len(reposTableKeybindings)] = common.Keybinding{Key: "?", ShortDesc: "Help"}
	ref := measure(m.renderKeybindingsLine(fullKbs))

	if ref > actual {
		return ref
	}
	return actual
}

// renderAlerts take list of alerts, calculate each alert y position and render it (it it's visible). Overlay each alert on top of main [view] (bg view)
func (m *Model) renderAlerts(
	view string,
	alertStates []alerts.AlertState,
) string {
	if len(alertStates) == 0 {
		return view
	}

	for _, alert := range alertStates {
		if alert.IsVisible {
			view = overlay.PlaceOverlay(
				alert.X,
				alert.Y,
				alert.AlertView,
				view,
				false,
				overlay.WithWhitespaceChars(" "),
			)
		}
	}
	return view
}
