package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckGitDirectory_WithGitDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0755))

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

func TestCheckGitDirectory_GitIsFile(t *testing.T) {
	dir := t.TempDir()
	// .git がファイル（worktree / submodule）の場合もリポジトリと見なす
	f, err := os.Create(filepath.Join(dir, ".git"))
	require.NoError(t, err)
	f.Close()

	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(orig)

	assert.NoError(t, checkGitDirectory())
}
