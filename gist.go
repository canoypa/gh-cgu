package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	gistDescription = "gh-cgu profiles"
)

// gistFileName returns the Gist filename based on the cached GitHub login ID
func gistFileName() string {
	return fmt.Sprintf("gh-cgu-%s-config.yml", ghLoginID)
}

// isBackgroundSyncCommand reports whether the current process is the internal
// background sync subprocess. Used to prevent recursive sync at startup.
func isBackgroundSyncCommand() bool {
	for _, arg := range os.Args[1:] {
		if arg == "_sync-gist" {
			return true
		}
	}
	return false
}

// findGistID searches all of the authenticated user's gists for one containing the config file
func findGistID() string {
	fileName := gistFileName()
	for page := 1; ; page++ {
		var gists []struct {
			ID    string                     `json:"id"`
			Files map[string]json.RawMessage `json:"files"`
		}
		if err := ghClient.Get(fmt.Sprintf("gists?per_page=100&page=%d", page), &gists); err != nil || len(gists) == 0 {
			break
		}
		for _, g := range gists {
			if _, ok := g.Files[fileName]; ok {
				return g.ID
			}
		}
		if len(gists) < 100 {
			break
		}
	}
	return ""
}

// resolveGistID returns the cached Gist ID from m, or searches the API if not cached.
// Updates m.GistID if found via API search.
func resolveGistID(m *syncMeta) string {
	if m.GistID != "" {
		return m.GistID
	}
	id := findGistID()
	if id != "" {
		m.GistID = id
	}
	return id
}

// pushConfigToGist synchronously pushes configFile to Gist (creates or updates).
// On success, m.GistID, m.LastSync, and m.LocalMtimeAtSync are updated.
func pushConfigToGist(configFile string, m *syncMeta) error {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	info, err := os.Stat(configFile)
	if err != nil {
		return err
	}
	localMtime := info.ModTime()

	type gistFile struct {
		Content string `json:"content"`
	}
	type gistPayload struct {
		Description string              `json:"description"`
		Public      bool                `json:"public"`
		Files       map[string]gistFile `json:"files"`
	}
	type gistResponse struct {
		ID        string    `json:"id"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	payload := gistPayload{
		Description: gistDescription,
		Public:      false,
		Files:       map[string]gistFile{gistFileName(): {Content: string(content)}},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	gistID := resolveGistID(m)
	var resp gistResponse
	if gistID == "" {
		if err := ghClient.Post("gists", bytes.NewReader(payloadBytes), &resp); err != nil {
			return err
		}
	} else {
		if err := ghClient.Patch(fmt.Sprintf("gists/%s", gistID), bytes.NewReader(payloadBytes), &resp); err != nil {
			return err
		}
	}

	if resp.ID != "" {
		m.GistID = resp.ID
	}
	if !resp.UpdatedAt.IsZero() {
		m.LastSync = resp.UpdatedAt
	}
	m.LocalMtimeAtSync = localMtime
	return nil
}

// pullConfigFromGist synchronously downloads Gist content and writes it to configFile.
// On success, m.GistID, m.LastSync, and m.LocalMtimeAtSync are updated.
func pullConfigFromGist(configFile string, m *syncMeta) error {
	gistID := resolveGistID(m)
	if gistID == "" {
		return fmt.Errorf("no Gist found")
	}

	var gist struct {
		ID        string    `json:"id"`
		UpdatedAt time.Time `json:"updated_at"`
		Files     map[string]struct {
			Content string `json:"content"`
		} `json:"files"`
	}
	if err := ghClient.Get(fmt.Sprintf("gists/%s", gistID), &gist); err != nil {
		return err
	}

	f, ok := gist.Files[gistFileName()]
	if !ok || f.Content == "" {
		return fmt.Errorf("config file not found in Gist")
	}

	if err := os.WriteFile(configFile, []byte(f.Content), 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	info, err := os.Stat(configFile)
	if err != nil {
		return err
	}

	m.GistID = gist.ID
	m.LastSync = gist.UpdatedAt
	m.LocalMtimeAtSync = info.ModTime()
	return nil
}

// syncAtStartup performs bidirectional sync between local config and Gist.
//   - If local was modified since last sync (e.g. offline edit) → push to Gist.
//   - Otherwise, check if Gist is newer → pull and reload viper.
//
// Network errors are silently ignored; local config is used as-is.
func syncAtStartup(configFile string) {
	if skipSyncForTesting || isBackgroundSyncCommand() {
		return
	}

	m := readSyncMeta()

	info, err := os.Stat(configFile)
	if err != nil {
		return
	}

	if m.LocalMtimeAtSync.IsZero() || info.ModTime().After(m.LocalMtimeAtSync) {
		// Local was modified since last sync (offline edit or never synced) → push
		if err := pushConfigToGist(configFile, &m); err != nil {
			return // Offline or error: silently continue with local
		}
		writeSyncMeta(m)
		return
	}

	// Local unchanged → check whether Gist has been updated on another machine.
	// Fetch the full Gist in one request so we already have the content if a pull is needed.
	gistID := resolveGistID(&m)
	if gistID == "" {
		return
	}

	var gist struct {
		ID        string    `json:"id"`
		UpdatedAt time.Time `json:"updated_at"`
		Files     map[string]struct {
			Content string `json:"content"`
		} `json:"files"`
	}
	if err := ghClient.Get(fmt.Sprintf("gists/%s", gistID), &gist); err != nil {
		return // Offline or error
	}

	if !gist.UpdatedAt.After(m.LastSync) {
		// Cache the GistID even when no pull is needed, so future startups skip findGistID.
		if m.GistID == "" {
			writeSyncMeta(m)
		}
		return
	}

	// Gist is newer → pull
	f, ok := gist.Files[gistFileName()]
	if !ok || f.Content == "" {
		return
	}
	if err := os.WriteFile(configFile, []byte(f.Content), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "! Failed to write config from Gist: %v\n", err)
		return
	}

	info, err = os.Stat(configFile)
	if err != nil {
		return
	}
	m.GistID = gist.ID
	m.LastSync = gist.UpdatedAt
	m.LocalMtimeAtSync = info.ModTime()
	writeSyncMeta(m)

	fmt.Println("✓ Profiles updated from Gist")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "! Failed to reload config after sync: %v\n", err)
	}
}

// syncToGist launches a detached background process to push config to Gist
func syncToGist() {
	if skipSyncForTesting {
		return
	}
	// Background sync relies on advisory file locking (Flock) which is not
	// implemented on Windows. Skip to avoid concurrent config corruption.
	if runtime.GOOS == "windows" {
		return
	}
	bin, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(bin, "_sync-gist")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = newSysProcAttr()
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "! failed to start background sync: %v\n", err)
	}
}

// doSyncToGist pushes the current config to Gist and updates sync meta.
// Called in the background subprocess (_sync-gist).
func doSyncToGist() {
	// Ensure only one background sync runs at a time.
	release, ok := acquireSyncLock()
	if !ok {
		return // Another background sync is already running
	}
	defer release()

	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return
	}
	m := readSyncMeta()
	if err := pushConfigToGist(configFile, &m); err != nil {
		fmt.Fprintf(os.Stderr, "! failed to sync profiles to Gist: %v\n", err)
		return
	}
	writeSyncMeta(m)
}

// pullFromGist downloads config from Gist and writes it to configFile.
// Used during first-run initialization. Updates sync meta on success.
func pullFromGist(configFile string) {
	m := readSyncMeta()
	if err := pullConfigFromGist(configFile, &m); err != nil {
		return // Silent: no Gist or offline
	}
	writeSyncMeta(m)
}

var syncGistCmd = &cobra.Command{
	Use:    "_sync-gist",
	Hidden: true,
	Args:   cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		doSyncToGist()
	},
}
