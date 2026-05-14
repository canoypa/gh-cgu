package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

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

// syncToGist launches a detached background process to push config to Gist
func syncToGist() {
	if skipSyncForTesting {
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
		fmt.Fprintf(os.Stderr, "warn: failed to start background sync: %v\n", err)
	}
}

// doSyncToGist pushes the current config to a Gist (creates if not exists)
func doSyncToGist() {
	configFile := viper.ConfigFileUsed()
	content, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	type gistFile struct {
		Content string `json:"content"`
	}
	type gistPayload struct {
		Description string              `json:"description"`
		Public      bool                `json:"public"`
		Files       map[string]gistFile `json:"files"`
	}

	payload := gistPayload{
		Description: gistDescription,
		Public:      false,
		Files:       map[string]gistFile{gistFileName(): {Content: string(content)}},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	gistID := findGistID()
	var result json.RawMessage
	var apiErr error
	if gistID == "" {
		apiErr = ghClient.Post("gists", bytes.NewReader(payloadBytes), &result)
	} else {
		apiErr = ghClient.Patch(fmt.Sprintf("gists/%s", gistID), bytes.NewReader(payloadBytes), &result)
	}
	if apiErr != nil {
		fmt.Fprintf(os.Stderr, "warn: failed to sync profiles to Gist: %v\n", apiErr)
	}
}

// pullFromGist downloads config from Gist and writes it to configFile
func pullFromGist(configFile string) {
	gistID := findGistID()
	if gistID == "" {
		return
	}

	var gist struct {
		Files map[string]struct {
			Content string `json:"content"`
		} `json:"files"`
	}
	if err := ghClient.Get(fmt.Sprintf("gists/%s", gistID), &gist); err != nil {
		return
	}

	f, ok := gist.Files[gistFileName()]
	if !ok || f.Content == "" {
		return
	}

	if err := os.WriteFile(configFile, []byte(f.Content), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "! Failed to write config from Gist: %v\n", err)
	}
}

var syncGistCmd = &cobra.Command{
	Use:    "_sync-gist",
	Hidden: true,
	Args:   cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		doSyncToGist()
	},
}
