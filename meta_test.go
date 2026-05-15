package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func withTempMetaFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.json")
	metaFileOverride = path
	t.Cleanup(func() { metaFileOverride = "" })
	return path
}

func TestReadSyncMeta_NoFile(t *testing.T) {
	withTempMetaFile(t)
	// File does not exist yet → should return zero-value struct, no panic.
	m := readSyncMeta()
	assert.Empty(t, m.GistID)
	assert.True(t, m.LastSync.IsZero())
	assert.True(t, m.LocalMtimeAtSync.IsZero())
}

func TestReadSyncMeta_CorruptYAML(t *testing.T) {
	path := withTempMetaFile(t)
	require.NoError(t, os.WriteFile(path, []byte("gist_id: [not valid yaml"), 0600))

	// Corrupt YAML → should return zero-value struct, no panic.
	m := readSyncMeta()
	assert.Empty(t, m.GistID)
	assert.True(t, m.LastSync.IsZero())
}

func TestWriteReadSyncMeta_Roundtrip(t *testing.T) {
	withTempMetaFile(t)

	ts := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	want := syncMeta{
		GistID:           "abc123",
		LastSync:         ts,
		LocalMtimeAtSync: ts.Add(time.Hour),
	}

	writeSyncMeta(want)
	got := readSyncMeta()

	assert.Equal(t, want.GistID, got.GistID)
	assert.Equal(t, want.LastSync.UTC(), got.LastSync.UTC())
	assert.Equal(t, want.LocalMtimeAtSync.UTC(), got.LocalMtimeAtSync.UTC())
}

func TestWriteSyncMeta_CreatesFile(t *testing.T) {
	path := withTempMetaFile(t)

	writeSyncMeta(syncMeta{GistID: "xyz"})

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var m syncMeta
	require.NoError(t, yaml.Unmarshal(data, &m))
	assert.Equal(t, "xyz", m.GistID)
}

func TestWriteSyncMeta_Overwrites(t *testing.T) {
	withTempMetaFile(t)

	writeSyncMeta(syncMeta{GistID: "first"})
	writeSyncMeta(syncMeta{GistID: "second"})

	m := readSyncMeta()
	assert.Equal(t, "second", m.GistID)
}
