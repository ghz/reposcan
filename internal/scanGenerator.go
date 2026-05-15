package internal

import (
	"time"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/scan"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func GenerateScanReport(
	configs config.Config,
) report.ScanReport {
	reportWarnings := []string{}

	// Find git repos at defined configs.Roots
	gitReposPaths, warnings := scan.FindGitRepos(configs.Roots, configs.DirIgnore)
	reportWarnings = append(reportWarnings, warnings...)

	allRepoStates, warnings := gitx.GetGitRepoStatesConcurrent(gitReposPaths, configs.MaxWorkers)
	reportWarnings = append(reportWarnings, warnings...)

	// filter repo states based on config OnlyFilter
	repoStates := make([]report.RepoState, 0, len(allRepoStates))
	for _, repoState := range allRepoStates {
		if filter(configs.Only, repoState) {
			repoStates = append(repoStates, repoState)
		}
	}

	allFolders, folderWarnings := scan.FindDirectSubdirs(configs.Roots, configs.DirIgnore)
	reportWarnings = append(reportWarnings, folderWarnings...)

	return report.ScanReport{
		Version:     configs.Version,
		GeneratedAt: time.Now(),
		RepoStates:  repoStates,
		AllFolders:  allFolders,
		Warnings:    reportWarnings,
	}
}

// GenerateFullScanReport is like GenerateScanReport but always returns all
// repositories regardless of the OnlyFilter in configs. Used by the TUI so
// it can manage its own view-mode filtering.
func GenerateFullScanReport(configs config.Config) report.ScanReport {
	full := configs
	full.Only = config.OnlyAll
	return GenerateScanReport(full)
}

// Filter repoState based on config only filter
// Returns true if repoState should be in output, false otherwise
func filter(f config.OnlyFilter, repoState report.RepoState) bool {
	switch f {
	case config.OnlyAll:
		return true
	case config.OnlyDirty:
		if repoState.IsDirty() {
			return true
		}
	case config.OnlyUncommitted:
		if len(repoState.UncommitedFiles) > 0 {
			return true
		}
	case config.OnlyUnpushed:
		return repoState.HaveUnpushedCommits()
	case config.OnlyUnpulled:
		return repoState.HaveUnpulledCommits()
	}

	return false
}
