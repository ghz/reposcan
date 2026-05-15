package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	rth "github.com/mabd-dev/reposcan/internal/render/tui/repostableheader"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// ViewMode controls which set of items is shown in the repos table.
type ViewMode int

const (
	ViewModeDirty       ViewMode = 0 // non-sync repos (default)
	ViewModeAllRepos    ViewMode = 1 // all git repos
	ViewModeAllDirs     ViewMode = 2 // all direct folders (repos + non-repos)
	ViewModeNonRepoDirs ViewMode = 3 // direct folders without a git repo
)

func (v ViewMode) Label() string {
	switch v {
	case ViewModeDirty:
		return "non-sync repos"
	case ViewModeAllRepos:
		return "all repos"
	case ViewModeAllDirs:
		return "all dirs"
	case ViewModeNonRepoDirs:
		return "non-repo dirs"
	}
	return ""
}

func (v ViewMode) Next() ViewMode {
	return (v + 1) % 4
}

func (v ViewMode) Prev() ViewMode {
	return (v + 3) % 4
}

type Model struct {
	// Loading stuff
	loading bool
	width   int
	height  int
	theme   theme.Theme

	// configs
	configs           config.Config
	reposBeingUpdated []string
	viewMode          ViewMode
	fullReport        report.ScanReport

	// Models
	reposTable          repostable.Model
	repoDetails         repodetails.Model
	rtHeader            rth.Header
	alerts              alerts.AlertModel
	reposFilter         textinput.Model
	createRepoNameInput textinput.Model
	createRepoFolderPath string
	createStep           createStep
	createKind           createKind

	focusStack []FocusState
}

// createStep tracks which step of the two-step "new repo" popup is active.
type createStep int

const (
	stepChooseKind createStep = iota // pick local / GitHub private / GitHub public
	stepEnterName                    // type the repo name and confirm
)

// createKind is the repo type chosen in the first step of the popup.
type createKind int

const (
	kindLocal createKind = iota
	kindGHPrivate
	kindGHPublic
)

func (k createKind) Label() string {
	switch k {
	case kindLocal:
		return "Local only"
	case kindGHPrivate:
		return "GitHub private"
	case kindGHPublic:
		return "GitHub public"
	}
	return ""
}

type createRepoResultMsg struct {
	label string // "local" | "GitHub private" | "GitHub public"
	err   error
}

func (m Model) IsReposFilterVisible() bool {
	return m.reposFilter.Focused() || len(strings.TrimSpace(m.reposFilter.Value())) != 0
}

type generateReport struct {
	configs config.Config
}

func (g *generateReport) Cmd() tea.Cmd {
	return func() tea.Msg {
		r := internal.GenerateFullScanReport(g.configs)
		return generateReportResponse{report: r}
	}
}

type generateReportResponse struct {
	report report.ScanReport
}
