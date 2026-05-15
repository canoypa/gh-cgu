package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateConfigIfNeeded_MovesFile(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "gh-cgu.yaml")
	newDir := filepath.Join(dir, "gh-cgu")
	newPath := filepath.Join(newDir, "config.yml")

	require.NoError(t, os.WriteFile(legacy, []byte("profiles:\n"), 0600))

	err := migrateConfigIfNeeded(legacy, newDir, newPath)
	require.NoError(t, err)

	// Legacy file should be gone.
	_, err = os.Stat(legacy)
	assert.True(t, os.IsNotExist(err), "legacy file should have been removed")

	// New file should exist with original content.
	data, err := os.ReadFile(newPath)
	require.NoError(t, err)
	assert.Equal(t, "profiles:\n", string(data))
}

func TestMigrateConfigIfNeeded_NoLegacy(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "gh-cgu.yaml")
	newDir := filepath.Join(dir, "gh-cgu")
	newPath := filepath.Join(newDir, "config.yml")

	// No legacy file → should be a no-op.
	err := migrateConfigIfNeeded(legacy, newDir, newPath)
	assert.NoError(t, err)

	_, err = os.Stat(newDir)
	assert.True(t, os.IsNotExist(err), "new dir should not have been created")
}

func TestMigrateConfigIfNeeded_NewAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "gh-cgu.yaml")
	newDir := filepath.Join(dir, "gh-cgu")
	newPath := filepath.Join(newDir, "config.yml")

	require.NoError(t, os.WriteFile(legacy, []byte("old content\n"), 0600))
	require.NoError(t, os.MkdirAll(newDir, 0700))
	require.NoError(t, os.WriteFile(newPath, []byte("new content\n"), 0600))

	// Both exist → no-op; legacy should be preserved, new should be untouched.
	err := migrateConfigIfNeeded(legacy, newDir, newPath)
	assert.NoError(t, err)

	data, err := os.ReadFile(legacy)
	require.NoError(t, err)
	assert.Equal(t, "old content\n", string(data), "legacy should be untouched")

	data, err = os.ReadFile(newPath)
	require.NoError(t, err)
	assert.Equal(t, "new content\n", string(data), "new path should be untouched")
}
