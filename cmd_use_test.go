package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false",
		"commit", "-q", "--allow-empty", "-m", "init")
	return dir
}

func TestCheckGitDirectory_WithGitDir(t *testing.T) {
	dir := initRepo(t)

	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(orig)

	assert.NoError(t, checkGitDirectory())
}

func TestCheckGitDirectory_WithoutGitDir(t *testing.T) {
	dir := t.TempDir()

	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(orig)

	assert.EqualError(t, checkGitDirectory(), "not a git repository")
}

func TestCheckGitDirectory_Subdirectory(t *testing.T) {
	dir := initRepo(t)
	sub := filepath.Join(dir, "sub")
	require.NoError(t, os.Mkdir(sub, 0755))

	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(sub))
	defer os.Chdir(orig)

	assert.NoError(t, checkGitDirectory())
}

func TestCheckGitDirectory_NestedWorktree(t *testing.T) {
	dir := initRepo(t)
	// メインチェックアウトの内側に置いた worktree（.git はファイル）
	wt := filepath.Join(dir, ".claude", "worktrees", "x")
	runGit(t, dir, "worktree", "add", "-q", wt)
	sub := filepath.Join(wt, "sub")
	require.NoError(t, os.Mkdir(sub, 0755))

	orig, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(orig)

	for _, d := range []string{wt, sub} {
		require.NoError(t, os.Chdir(d))
		assert.NoError(t, checkGitDirectory(), d)
	}
}
