package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mabd-dev/reposcan/pkg/report"
)

// FindGitRepos walks each root and returns directories that look like Git worktrees.
// Simple rules:
// - A directory containing `.git` (directory) is a repo root.
// - Or a `.git` file whose contents include "gitdir:" (worktrees/submodules).
// - When we find a repo root, we SkipDir to avoid descending into nested repos (for now).
func FindGitRepos(
	roots []string,
	dirignore []string,
) (gitReposPaths []string, warnings []string) {
	matcher := NewIgnoreMatcher(roots, dirignore)

	visitedDir := map[string]struct{}{}

	for _, root := range roots {
		root = os.ExpandEnv(root)

		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// possible errors: permission denied
				warnings = append(warnings, err.Error())
				return nil
			}

			if !d.IsDir() {
				return nil
			}

			if matcher.ShouldIgnore(path) {
				return fs.SkipDir
			}

			if _, visited := visitedDir[path]; visited {
				return fs.SkipDir
			}
			visitedDir[path] = struct{}{}

			if isGitRepo(path) {
				gitReposPaths = append(gitReposPaths, path)
				return fs.SkipDir
			}
			return nil
		})
	}

	return removeDuplicates(gitReposPaths), warnings
}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if file, err := os.Lstat(gitPath); err == nil {
		if file.IsDir() {
			return true
		} else {
			// git worktrees and submodules use a .git file containing "gitdir: ..."
			b, err := os.ReadFile(gitPath)
			if err != nil {
				return false
			}
			return strings.Contains(string(b), "gitdir:")
		}
	}
	return false
}

// FindDirectSubdirs returns all immediate subdirectories of each root that are
// not ignored, marking whether each is a Git repository.
func FindDirectSubdirs(roots []string, dirignore []string) ([]report.FolderEntry, []string) {
	matcher := NewIgnoreMatcher(roots, dirignore)
	var entries []report.FolderEntry
	var warnings []string
	seen := map[string]struct{}{}

	for _, root := range roots {
		root = os.ExpandEnv(root)
		dirEntries, err := os.ReadDir(root)
		if err != nil {
			warnings = append(warnings, err.Error())
			continue
		}
		for _, de := range dirEntries {
			if !de.IsDir() {
				continue
			}
			p := filepath.Join(root, de.Name())
			if matcher.ShouldIgnore(p) {
				continue
			}
			if _, visited := seen[p]; visited {
				continue
			}
			seen[p] = struct{}{}
			entries = append(entries, report.FolderEntry{
				Path:   p,
				Name:   de.Name(),
				IsRepo: isGitRepo(p),
			})
		}
	}
	return entries, warnings
}

func removeDuplicates(strs []string) []string {
	seen := make(map[string]struct{}, len(strs))
	distinct := make([]string, 0, len(strs))

	for _, s := range strs {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			distinct = append(distinct, s)
		}
	}
	return distinct
}
