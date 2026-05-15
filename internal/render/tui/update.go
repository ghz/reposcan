package tui

import (
	"errors"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"golang.design/x/clipboard"
)

var (
	deleteRepoTimeout = 3 * time.Second
	removeAll         = os.RemoveAll
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Async results, ticks and window-size changes are global: they must be
	// handled no matter which model currently has focus. Otherwise an
	// operation in flight (a scan, a delete, a git push...) can get stuck --
	// e.g. m.loading never clears and the UI freezes on "Loading..." -- if the
	// user opens a popup while it runs.
	if _, isKey := msg.(tea.KeyMsg); !isKey {
		if nm, cmd := defaultUpdate(m, msg); nm != nil {
			return nm, cmd
		}
	}

	// While a scan is in flight the view only shows "Loading...", so action
	// keys would open popups over an invisible UI. Ignore everything except
	// quit until the scan result clears m.loading.
	if m.loading {
		if key, isKey := msg.(tea.KeyMsg); isKey {
			if s := keyString(key); s == "q" || s == "esc" || s == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	switch m.currentFocus() {
	case FocusReposTable:
		return m.updateReposTable(msg)
	case FocusReposFilter:
		return m.updateReposFilter(msg)
	case FocusHelpPopup:
		return m.keybindingPopup(msg)
	case FocusCreateRepoPopup:
		return m.updateCreateRepoPopup(msg)
	case FocusGitMenuPopup:
		return m.updateGitMenuPopup(msg)
	case FocusDeleteRepoPopup:
		return m.updateDeleteRepoPopup(msg)
	}
	return m, nil
}

func (m Model) updateReposTable(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keyString(msg) {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "g", "G":
			if m.reposTable.GetCurrentRepoState() == nil {
				return m, nil
			}
			m.pushFocus(FocusGitMenuPopup)
			return m, nil
		case "w":
			rs := m.reposTable.GetCurrentRepoState()
			if rs == nil {
				return m, nil
			}
			return m, openRemoteForRepo(rs.Path)
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
				cmd := commandForOpenPath(editor, path)
				_ = cmd.Start()
			}
			return m, nil
		case "e":
			path := m.reposTable.GetCurrentPath()
			if path == "" {
				return m, nil
			}
			if err := openFileManager(path); err != nil {
				return m, makeAlert(alerts.MsgTypeError, "open failed: "+err.Error())
			}
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
			if m.reposTable.GetCurrentPath() == "" {
				return m, nil
			}
			m.deleteConfirmInput.SetValue("")
			m.pushFocus(FocusDeleteRepoPopup)
			return m, nil
		case "tab":
			m.repoDetails.ToggleSubMode(m.reposTable.GetCurrentRepoState())
			return m, nil
		case "c":
			path := m.reposTable.GetCurrentPath()
			if path == "" {
				return m, nil
			}

			clipboard.Write(clipboard.FmtText, []byte(shellEscapePath(path)))
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

	prevCursor := m.reposTable.Cursor()
	var cmd tea.Cmd
	m.reposTable, cmd = m.reposTable.Update(msg)
	if m.reposTable.Cursor() != prevCursor && m.repoDetails.SubMode() == repodetails.DetailsSubModeCommits {
		m.repoDetails.RefetchCommits(m.reposTable.GetCurrentRepoState())
	}
	return m, cmd
}

func (m Model) updateGitMenuPopup(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		return m, nil
	}

	rs := m.reposTable.GetCurrentRepoState()
	if rs == nil {
		m.popFocus(false)
		return m, nil
	}

	switch keyString(keyMsg) {
	case "q", "esc", "ctrl+c":
		m.popFocus(false)
		return m, nil
	case "1":
		m.popFocus(false)
		return m, quickSaveCmd(rs.Path)
	case "2":
		m.popFocus(false)
		return m, gitPush(m)
	case "3":
		m.popFocus(false)
		return m, gitPull(m)
	case "4":
		m.popFocus(false)
		return m, gitFetch(m)
	case "5":
		m.popFocus(false)
		return m, openRemoteForRepo(rs.Path)
	}

	return m, nil
}

func openRemoteForRepo(repoPath string) tea.Cmd {
	url, err := gitx.GetRemoteWebURL(repoPath)
	if err != nil || url == "" {
		return makeAlert(alerts.MsgTypeError, "remote URL unavailable")
	}
	openBrowser(url)
	return nil
}

func (m Model) updateDeleteRepoPopup(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		var cmd tea.Cmd
		m.deleteConfirmInput, cmd = m.deleteConfirmInput.Update(msg)
		return m, cmd
	}

	switch keyString(keyMsg) {
	case "esc", "ctrl+c":
		m.popFocus(true)
		return m, nil
	case "enter":
		targetName, path, _ := m.deleteTarget()
		if strings.TrimSpace(path) == "" {
			m.popFocus(true)
			return m, nil
		}
		if strings.TrimSpace(m.deleteConfirmInput.Value()) != "YES" {
			return m, makeAlert(alerts.MsgTypeError, "Type YES to confirm deletion")
		}
		m.popFocus(true)
		return m, deleteRepoCmd(targetName, path)
	}

	var cmd tea.Cmd
	m.deleteConfirmInput, cmd = m.deleteConfirmInput.Update(msg)
	return m, cmd
}

func (m Model) updateReposFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keyString(msg) {
		case "esc", "ctrl+c":
			m.popFocus(true)

			return m, nil
		case "enter":
			m.acceptReposFilter()
			return m, nil
		}
	}

	// Update the repos list on each keystroke.
	var cmd tea.Cmd
	m.reposFilter, cmd = m.reposFilter.Update(msg)

	m.reposTable.Filter(m.reposFilter.Value())

	return m, cmd
}

func (m *Model) acceptReposFilter() {
	selectedRepoID := ""
	if rs := m.reposTable.GetCurrentRepoState(); rs != nil {
		selectedRepoID = rs.ID
	}

	m.reposFilter.SetValue("")
	m.reposTable.Filter("")
	m.reposTable.SetCursorByRepoID(selectedRepoID)
	m.popFocus(false)
}

func (m Model) keybindingPopup(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keyString(msg) {
		case "q", "esc":
			m.popFocus(true)
			return m, nil
		}
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
		switch keyString(keyMsg) {
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
	switch keyString(keyMsg) {
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

func keyString(msg tea.KeyMsg) string {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		return string(msg.Runes[0])
	}
	return msg.String()
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
		err := rollbackCreatedGitDirOnError(path, func() error {
			if err := gitx.InitRepo(path); err != nil {
				return err
			}
			if err := gitx.AddAll(path); err != nil {
				return err
			}
			_, err := gitx.CommitInitialIfNeeded(path)
			return err
		})
		if err != nil {
			return createRepoResultMsg{label: "local", err: err}
		}
		return createRepoResultMsg{label: "local"}
	}
}

func createGitHubRepoCmd(name, path string, private bool) tea.Cmd {
	label := "GitHub public"
	if private {
		label = "GitHub private"
	}
	return func() tea.Msg {
		err := rollbackCreatedGitDirOnError(path, func() error {
			if err := gitx.InitRepo(path); err != nil {
				return err
			}
			if err := gitx.AddAll(path); err != nil {
				return err
			}
			committed, err := gitx.CommitInitialIfNeeded(path)
			if err != nil {
				return err
			}
			return gitx.GitHubCreateRepo(name, path, private, committed)
		})
		if err != nil {
			return createRepoResultMsg{label: label, err: err}
		}
		return createRepoResultMsg{label: label}
	}
}

func deleteRepoCmd(targetName, path string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(path) == "" {
			return deleteRepoResultMsg{repoName: targetName, err: errors.New("empty folder path")}
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- removeAll(path)
		}()

		select {
		case err := <-errCh:
			return deleteRepoResultMsg{repoName: targetName, err: err}
		case <-time.After(deleteRepoTimeout):
			return deleteRepoResultMsg{repoName: targetName, err: errors.New("delete timed out; the folder may be locked by another process")}
		}
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
		return m, tea.Batch(makeAlert(alerts.AlertTypeInfo, "Repo created ("+msg.label+")"), request.Cmd())

	case deleteRepoResultMsg:
		if msg.err != nil {
			return m, makeAlert(alerts.MsgTypeError, "delete failed: "+msg.err.Error())
		}
		m.loading = true
		request := generateReport{configs: m.configs}
		label := msg.repoName
		if label == "" {
			label = "folder"
		}
		return m, tea.Batch(makeAlert(alerts.AlertTypeInfo, "Folder deleted: "+label), request.Cmd())

	case alerts.AddAlertMsg, alerts.TickMsg:
		var cmd tea.Cmd
		m.alerts, cmd = m.alerts.Update(msg)
		return m, cmd
	}

	return nil, nil
}
