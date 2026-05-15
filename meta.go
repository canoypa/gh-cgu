package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// syncMeta holds bidirectional sync state persisted between invocations.
type syncMeta struct {
	// GistID caches the Gist ID to avoid full-page search on every startup.
	GistID string `yaml:"gist_id,omitempty"`
	// LastSync is the Gist's updated_at at the time of the last successful sync.
	// Used to detect whether the remote Gist has been updated since last sync.
	LastSync time.Time `yaml:"last_sync,omitempty"`
	// LocalMtimeAtSync is the local config file's mtime at the time of the last
	// successful sync. Used to detect local edits without relying on server clock.
	LocalMtimeAtSync time.Time `yaml:"local_mtime_at_sync,omitempty"`
}

func metaFilePath() string {
	if metaFileOverride != "" {
		return metaFileOverride
	}
	homePath, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homePath, ".config", "gh-cgu", "meta.yml")
}

func syncLockPath() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homePath, ".config", "gh-cgu", "sync.lock")
}

func readSyncMeta() syncMeta {
	path := metaFilePath()
	if path == "" {
		return syncMeta{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "! failed to read sync metadata: %v\n", err)
		}
		return syncMeta{}
	}
	var m syncMeta
	if err := yaml.Unmarshal(data, &m); err != nil {
		fmt.Fprintf(os.Stderr, "! sync metadata is corrupt, resetting: %v\n", err)
		return syncMeta{}
	}
	return m
}

func writeSyncMeta(m syncMeta) {
	path := metaFilePath()
	if path == "" {
		return
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "! failed to save sync state: %v\n", err)
	}
}
