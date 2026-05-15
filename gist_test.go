package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBackgroundSyncCommand_False(t *testing.T) {
	// Restore os.Args after test.
	orig := os.Args
	t.Cleanup(func() { os.Args = orig })

	os.Args = []string{"gh-cgu", "add", "Alice", "alice@example.com"}
	assert.False(t, isBackgroundSyncCommand())
}

func TestIsBackgroundSyncCommand_True(t *testing.T) {
	orig := os.Args
	t.Cleanup(func() { os.Args = orig })

	os.Args = []string{"gh-cgu", "_sync-gist"}
	assert.True(t, isBackgroundSyncCommand())
}

func TestIsBackgroundSyncCommand_TrueAmongOtherArgs(t *testing.T) {
	orig := os.Args
	t.Cleanup(func() { os.Args = orig })

	os.Args = []string{"gh-cgu", "--flag", "_sync-gist"}
	assert.True(t, isBackgroundSyncCommand())
}

func TestSyncAtStartup_SkipsWhenTesting(t *testing.T) {
	// syncAtStartup must be a no-op when skipSyncForTesting is true.
	// We verify no panic occurs even with an empty configFile.
	orig := skipSyncForTesting
	skipSyncForTesting = true
	t.Cleanup(func() { skipSyncForTesting = orig })

	// Should return immediately without touching the file or network.
	assert.NotPanics(t, func() {
		syncAtStartup("/nonexistent/path/config.yml")
	})
}

func TestSyncAtStartup_SkipsWhenBackgroundSync(t *testing.T) {
	origArgs := os.Args
	origSkip := skipSyncForTesting
	t.Cleanup(func() {
		os.Args = origArgs
		skipSyncForTesting = origSkip
	})

	skipSyncForTesting = false
	os.Args = []string{"gh-cgu", "_sync-gist"}

	// Should return immediately because isBackgroundSyncCommand() is true.
	assert.NotPanics(t, func() {
		syncAtStartup("/nonexistent/path/config.yml")
	})
}
