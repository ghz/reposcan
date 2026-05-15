package tui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRollbackCreatedGitDirOnErrorRemovesNewGitDir(t *testing.T) {
	dir := t.TempDir()
	wantErr := errors.New("gh failed")

	err := rollbackCreatedGitDirOnError(dir, func() error {
		if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
			t.Fatalf("Mkdir(.git) error = %v", err)
		}
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("rollbackCreatedGitDirOnError() error = %v, want %v", err, wantErr)
	}
	if _, statErr := os.Lstat(filepath.Join(dir, ".git")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf(".git still exists or unexpected stat error: %v", statErr)
	}
}

func TestRollbackCreatedGitDirOnErrorKeepsExistingGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatalf("Mkdir(.git) error = %v", err)
	}

	wantErr := errors.New("add failed")
	err := rollbackCreatedGitDirOnError(dir, func() error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("rollbackCreatedGitDirOnError() error = %v, want %v", err, wantErr)
	}
	if _, statErr := os.Lstat(gitDir); statErr != nil {
		t.Fatalf(".git was removed, stat error = %v", statErr)
	}
}

func TestRollbackCreatedGitDirOnSuccessKeepsNewGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")

	err := rollbackCreatedGitDirOnError(dir, func() error {
		return os.Mkdir(gitDir, 0o755)
	})

	if err != nil {
		t.Fatalf("rollbackCreatedGitDirOnError() error = %v", err)
	}
	if _, statErr := os.Lstat(gitDir); statErr != nil {
		t.Fatalf(".git was removed after success, stat error = %v", statErr)
	}
}

func TestCreateGitHubRepoCmdRemovesGitDirWhenCreateFails(t *testing.T) {
	dir := t.TempDir()

	msg := createGitHubRepoCmd("", dir, false)()
	result, ok := msg.(createRepoResultMsg)
	if !ok {
		t.Fatalf("message = %T, want createRepoResultMsg", msg)
	}
	if result.err == nil {
		t.Fatal("createGitHubRepoCmd() error = nil, want error")
	}
	if _, statErr := os.Lstat(filepath.Join(dir, ".git")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf(".git still exists or unexpected stat error: %v", statErr)
	}
}
