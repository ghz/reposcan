package tui

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"golang.design/x/clipboard"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentFocus() {
	case FocusReposTable:
		return m.updateReposTable(msg)
	case FocusReposFilter:
		return m.updateReposFilter(msg)
	case FocusHelpPopup:
		return m.keybindingPopup(msg)
	case FocusCreateRepoPopup:
		return m.updateCreateRepoPopup(msg)
	}
	return m, nil
}

func (m Model) updateReposTable(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "S":
			rs := m.reposTable.GetCurrentRepoState()
			if rs == nil {
				return m, nil
			}
			return m, quickSaveCmd(rs.Path)
		case "p":
			return m, gitPull(m)
		case "P":
			return m, gitPush(m)
		case "F":
			return m, gitFetch(m)
		case "right", "l":
			m.viewMode = m.viewMode.Next()
			m.applyViewMode()
			return m, nil
		case "left", "h":
			m.viewMode = m.viewMode.Prev()
			m.applyViewMode()
			return m, nil
		case "o":
			path := m.reposTable.GetCurrentPath()
			if path != "" {
				editor := m.configs.Editor
				if editor == "" {
					editor = "code"
				}
				cmd := exec.Command(editor, path)
				_ = cmd.Start()
			}
			return m, nil
		case "g":
			rs := m.reposTable.GetCurrentRepoState()
			if rs == nil {
				return m, nil
			}
			url, err := gitx.GetRemoteWebURL(rs.Path)
			if err != nil || url == "" {
				return m, nil
			}
			openBrowser(url)
			return m, nil
		case "f":
			rs := m.reposTable.GetCurrentRepoState()
			if rs == nil {
				return m, nil
			}
			m.toggleFavorite(rs.Path)
			return m, nil
		case "n":
			entry := m.reposTable.GetCurrentFolderEntry()
			if entry == nil || entry.IsRepo {
				return m, nil
			}
			m.createRepoFolderPath = entry.Path
			m.createRepoNameInput.SetValue(entry.Name)
			m.createStep = stepChooseKind
			m.pushFocus(FocusCreateRepoPopup)
			return m, nil
		case "d":
			m.repoDetails.ToggleSubMode(m.reposTable.GetCurrentRepoState())
			return m, nil
		case "c":
			rs := m.reposTable.GetCurrentRepoState()
			if rs == nil {
				return m, nil
			}

			path := shellEscapePath(rs.Path)
			clipboard.Write(clipboard.FmtText, []byte(path))
			return m, makeAlert(alerts.AlertTypeInfo, "Path copied to clipboard")
		case "r":
			m.loading = true
			request := generateReport{configs: m.configs}
			return m, request.Cmd()
		case "/":
			m.pushFocus(FocusReposFilter)
			return m, nil
		case "?":
			m.pushFocus(FocusHelpPopup)
			return m, nil
		}
	}

	var cmd tea.Cmd
	nm, cmd := defaultUpdate(m, msg)

	if nm != nil {
		return nm, cmd
	}

	prevCursor := m.reposTable.Cursor()
	m.reposTable, cmd = m.reposTable.Update(msg)
	if m.reposTable.Cursor() != prevCursor && m.repoDetails.SubMode() == repodetails.DetailsSubModeCommits {
		m.repoDetails.RefetchCommits(m.reposTable.GetCurrentRepoState())
	}
	return m, cmd
}

func (m Model) updateReposFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.popFocus(true)

			return m, nil
		case "enter":
			emptyQuery := len(strings.TrimSpace(m.reposFilter.Value())) == 0

			m.popFocus(emptyQuery)

			return m, nil
		}
	}

	// on each keystorke, update repos list
	var cmd tea.Cmd
	m.reposFilter, cmd = m.reposFilter.Update(msg)

	m.reposTable.Filter(m.reposFilter.Value())

	return m, cmd
}

func (m Model) keybindingPopup(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.popFocus(true)
			return m, nil
		}
	}

	var cmd tea.Cmd
	nm, cmd := defaultUpdate(m, msg)

	if nm != nil {
		return nm, cmd
	}
	return m, nil
}

func (m Model) updateCreateRepoPopup(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		var cmd tea.Cmd
		m.createRepoNameInput, cmd = m.createRepoNameInput.Update(msg)
		return m, cmd
	}

	if m.createStep == stepChooseKind {
		switch keyMsg.String() {
		case "esc", "ctrl+c":
			m.popFocus(true)
		case "1":
			m.createKind = kindLocal
			m.createStep = stepEnterName
		case "2":
			m.createKind = kindGHPrivate
			m.createStep = stepEnterName
		case "3":
			m.createKind = kindGHPublic
			m.createStep = stepEnterName
		}
		return m, nil
	}

	// stepEnterName
	switch keyMsg.String() {
	case "ctrl+c":
		m.popFocus(true)
		return m, nil
	case "esc":
		m.createStep = stepChooseKind
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.createRepoNameInput.Value())
		if name == "" {
			return m, nil
		}
		path := m.createRepoFolderPath
		kind := m.createKind
		m.popFocus(true)
		switch kind {
		case kindLocal:
			return m, createLocalRepoCmd(path)
		case kindGHPrivate:
			return m, createGitHubRepoCmd(name, path, true)
		case kindGHPublic:
			return m, createGitHubRepoCmd(name, path, false)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.createRepoNameInput, cmd = m.createRepoNameInput.Update(msg)
	return m, cmd
}

type quickSaveResultMsg struct {
	result gitx.QuickSaveResult
	err    error
}

func quickSaveCmd(path string) tea.Cmd {
	return func() tea.Msg {
		result, err := gitx.QuickSave(path)
		return quickSaveResultMsg{result: result, err: err}
	}
}

func createLocalRepoCmd(path string) tea.Cmd {
	return func() tea.Msg {
		if err := gitx.InitRepo(path); err != nil {
			return createRepoResultMsg{label: "local", err: err}
		}
		if err := gitx.AddAll(path); err != nil {
			return createRepoResultMsg{label: "local", err: err}
		}
		if err := gitx.CommitInitial(path); err != nil {
			return createRepoResultMsg{label: "local", err: err}
		}
		return createRepoResultMsg{label: "local"}
	}
}

func createGitHubRepoCmd(name, path string, private bool) tea.Cmd {
	label := "GitHub public"
	if private {
		label = "GitHub privé"
	}
	return func() tea.Msg {
		if err := gitx.InitRepo(path); err != nil {
			return createRepoResultMsg{label: label, err: err}
		}
		if err := gitx.AddAll(path); err != nil {
			return createRepoResultMsg{label: label, err: err}
		}
		if err := gitx.CommitInitial(path); err != nil {
			return createRepoResultMsg{label: label, err: err}
		}
		if err := gitx.GitHubCreateRepo(name, path, private); err != nil {
			return createRepoResultMsg{label: label, err: err}
		}
		return createRepoResultMsg{label: label}
	}
}

func defaultUpdate(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case quickSaveResultMsg:
		if msg.err != nil {
			return m, makeAlert(alerts.MsgTypeError, msg.err.Error())
		}
		label := "pushed"
		if msg.result.Committed {
			label = "committed (wip) + pushed"
		}
		return m, tea.Batch(makeAlert(alerts.AlertTypeInfo, label), gitRefreshRepo(m))

	case gitPushResultMsg:
		if len(msg.Err) != 0 {
			logger.Warn(msg.Err)
			return m, makeAlert(alerts.MsgTypeError, "push failed: "+msg.Err)
		}
		return m, tea.Batch(makeAlert(alerts.AlertTypeInfo, "pushed"), gitRefreshRepo(m))

	case gitPullResultMsg:
		if len(msg.Err) != 0 {
			logger.Warn(msg.Err)
			return m, makeAlert(alerts.MsgTypeError, "pull failed: "+msg.Err)
		}
		rs := m.reposTable.GetCurrentRepoState()
		if rs == nil {
			return m, nil
		}
		index := getRepoIndex(m.reposBeingUpdated, rs.ID)
		if index != -1 {
			m.reposBeingUpdated = deleteRepo(m.reposBeingUpdated, index)
		}
		return m, tea.Batch(makeAlert(alerts.AlertTypeInfo, "pulled"), gitRefreshRepo(m))

	case gitFetchResultMsg:
		if len(msg.Err) != 0 {
			logger.Warn(msg.Err)
			return m, makeAlert(alerts.MsgTypeError, "fetch failed: "+msg.Err)
		}
		rs := m.reposTable.GetCurrentRepoState()
		if rs == nil {
			return m, nil
		}
		index := getRepoIndex(m.reposBeingUpdated, rs.ID)
		if index != -1 {
			m.reposBeingUpdated = deleteRepo(m.reposBeingUpdated, index)
		}
		return m, tea.Batch(makeAlert(alerts.AlertTypeInfo, "fetched"), gitRefreshRepo(m))

	case gitRefreshRepoResultMsg:
		m.reposTable.UpdateRepoState(msg.index, msg.newRepoState)

		return m, nil
	case generateReportResponse:
		m.loading = false
		m.fullReport = msg.report
		m.applyViewMode()
		return m, nil

	case createRepoResultMsg:
		if msg.err != nil {
			return m, makeAlert(alerts.MsgTypeError, msg.err.Error())
		}
		m.loading = true
		request := generateReport{configs: m.configs}
		return m, tea.Batch(makeAlert(alerts.AlertTypeInfo, "Repo créé ("+msg.label+")"), request.Cmd())

	case alerts.AddAlertMsg, alerts.TickMsg:
		var cmd tea.Cmd
		m.alerts, cmd = m.alerts.Update(msg)
		return m, cmd
	}

	return nil, nil
}
