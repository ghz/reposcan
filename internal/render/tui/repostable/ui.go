package repostable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

const (
	RepoW        = 30
	BranchW      = 30
	RemoteStateW = 40
)

func createColumns(maxWidth int) []table.Column {
	repoW := maxWidth * RepoW / 100
	branchW := maxWidth * BranchW / 100
	remoteStateW := maxWidth * RemoteStateW / 100

	return []table.Column{
		{Title: "Repo", Width: repoW},
		{Title: "Branch", Width: branchW},
		{Title: "State", Width: remoteStateW},
	}
}

func createRows(repoStates []report.RepoState, theme theme.Theme) []table.Row {
	rows := make([]table.Row, 0, len(repoStates))
	for _, rs := range repoStates {
		state := getStateColumnStr(rs, theme)

		rows = append(rows, table.Row{
			rs.Repo,
			rs.Branch,
			state,
		})
	}
	return rows
}

func getStateColumnStr(rs report.RepoState, t theme.Theme) string {
	uc := len(rs.UncommitedFiles)

	ucColor := t.Colors.Muted
	if uc > 0 {
		ucColor = t.Colors.Error
	}
	ucStr := lipgloss.NewStyle().Foreground(ucColor).Render(fmt.Sprintf("⏳%-d", uc)) + " "

	parts := []string{}
	for _, remoteStatus := range rs.RemoteStatus {
		var statusParts []string

		if remoteStatus.Ahead < 0 {
			statusParts = append(statusParts, lipgloss.NewStyle().Foreground(t.Colors.Error).Render("x"))
		} else {
			aheadColor := t.Colors.Muted
			if remoteStatus.Ahead > 0 {
				aheadColor = t.Colors.Warning
			}
			statusParts = append(statusParts, lipgloss.NewStyle().Foreground(aheadColor).Render(fmt.Sprintf("↑%-d", remoteStatus.Ahead)))
		}

		if remoteStatus.Behind < 0 {
			statusParts = append(statusParts, lipgloss.NewStyle().Foreground(t.Colors.Error).Render("x"))
		} else {
			behindColor := t.Colors.Muted
			if remoteStatus.Behind > 0 {
				behindColor = t.Colors.Warning
			}
			statusParts = append(statusParts, lipgloss.NewStyle().Foreground(behindColor).Render(fmt.Sprintf("↓%-d", remoteStatus.Behind)))
		}

		if remoteStatus.Remote != "" && !(len(rs.RemoteStatus) == 1 && remoteStatus.Remote == "origin") {
			remoteName := t.Styles.Base.Render(fmt.Sprintf("(%s)", remoteStatus.Remote))
			statusParts = append(statusParts, remoteName)
		}

		parts = append(parts, strings.Join(statusParts, " "))
	}

	s := ucStr
	s += strings.Join(parts, " | ")

	return s
}

func setKeymaps(km table.KeyMap) {
	km.LineUp.SetKeys("up", "k")
	km.LineDown.SetKeys("down", "j")
	km.PageUp.SetKeys("pgup", tea.KeyCtrlU.String())
	km.PageDown.SetKeys("pgdn", tea.KeyCtrlD.String())
	km.GotoTop.SetKeys("home", "g")
	km.GotoBottom.SetKeys("end", "G")
}

func getRepoIndex(repos []report.RepoState, id string) int {
	for i, s := range repos {
		if s.ID == id {
			return i
		}
	}
	return -1
}
