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

	tabBar := m.tabBarView()
	tabBarHeight := lipgloss.Height(tabBar)

	bodyHeight := m.height - footerHeight - tabBarHeight

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

	view := lipgloss.JoinVertical(lipgloss.Left, tabBar, body, footer)

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

// tabBarView renders a horizontal bar with every ViewMode, highlighting the
// active one. It mirrors what the "tab" key cycles through.
func (m *Model) tabBarView() string {
	modes := []ViewMode{
		ViewModeDirty,
		ViewModeAllRepos,
		ViewModeAllDirs,
		ViewModeNonRepoDirs,
	}

	tabs := make([]string, 0, len(modes))
	for _, mode := range modes {
		label := mode.Label()
		if label != "" {
			label = strings.ToUpper(label[:1]) + label[1:]
		}

		style := m.theme.Styles.Base.Padding(0, 1)
		if mode == m.viewMode {
			style = style.
				Foreground(m.theme.Colors.Background).
				Background(m.theme.Colors.Accent).
				Bold(true)
		} else {
			style = style.Foreground(m.theme.Colors.Muted)
		}
		tabs = append(tabs, style.Render(label))
	}

	tabsBlock := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	hint := m.theme.Styles.Muted.Render("← → switch view ")

	gap := m.width - lipgloss.Width(tabsBlock) - lipgloss.Width(hint)
	if gap < 1 {
		gap = 1
	}

	bar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tabsBlock,
		strings.Repeat(" ", gap),
		hint,
	)
	return lipgloss.NewStyle().Width(m.width).MaxWidth(m.width).Render(bar)
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
