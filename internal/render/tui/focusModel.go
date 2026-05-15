package tui

import (
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
)

type FocusState int

const (
	FocusReposTable FocusState = iota
	FocusReposFilter
	FocusHelpPopup
	FocusCreateRepoPopup
)

func (m Model) currentFocus() FocusState {
	if len(m.focusStack) == 0 {
		return FocusReposTable
	}
	return m.focusStack[len(m.focusStack)-1]
}

func (m *Model) pushFocus(state FocusState) {
	m.blurCurrentModel()
	m.focusStack = append(m.focusStack, state)
	m.focusCurrentModel()
}

func (m *Model) popFocus(reset bool) Model {
	m.blurCurrentModel()
	if reset {
		m.resetCurrentModel()
	}

	if len(m.focusStack) > 1 {
		m.focusStack = m.focusStack[:len(m.focusStack)-1]
	}

	m.focusCurrentModel()
	if reset {
		m.resetCurrentModel()
	}

	return *m
}

func (m *Model) focusCurrentModel() {
	switch m.currentFocus() {
	case FocusReposTable:
		m.reposTable.Focus()
	case FocusReposFilter:
		m.reposFilter.Focus()
	case FocusCreateRepoPopup:
		m.createRepoNameInput.Focus()
	case FocusHelpPopup:
		break
	}
}

func (m *Model) blurCurrentModel() {
	switch m.currentFocus() {
	case FocusReposTable:
		m.reposTable.Blur()
	case FocusReposFilter:
		m.reposFilter.Blur()
	case FocusCreateRepoPopup:
		m.createRepoNameInput.Blur()
	case FocusHelpPopup:
		break
	}
}

func (m *Model) resetCurrentModel() {
	switch m.currentFocus() {
	case FocusReposTable:
		m.reposTable.Filter("")
	case FocusReposFilter:
		m.reposFilter.SetValue("")
	case FocusCreateRepoPopup:
		m.createRepoNameInput.SetValue("")
	case FocusHelpPopup:
		break
	}
}

func (m *Model) keybindings() []common.Keybinding {
	switch m.currentFocus() {
	case FocusReposTable:
		entry := m.reposTable.GetCurrentFolderEntry()
		if entry != nil && !entry.IsRepo {
			return reposTableKeybindings
		}
		result := make([]common.Keybinding, 0, len(reposTableKeybindings))
		for _, kb := range reposTableKeybindings {
			if kb.Key != "n" {
				result = append(result, kb)
			}
		}
		return result
	case FocusReposFilter:
		return reposTableFilterKeybindings
	case FocusHelpPopup:
		return helpPopupKeybindings
	case FocusCreateRepoPopup:
		if m.createStep == stepEnterName {
			return createRepoNameKeybindings
		}
		return createRepoKindKeybindings
	}
	return []common.Keybinding{}
}
