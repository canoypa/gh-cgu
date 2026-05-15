package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ghLoginID string
	ghClient  *api.RESTClient
)

// migrateConfigIfNeeded moves the legacy config file to the new location.
// It is a no-op if the legacy file does not exist or the new location already has a file.
func migrateConfigIfNeeded(legacyPath, newDir, newPath string) error {
	if _, err := os.Stat(legacyPath); err != nil {
		return nil // no legacy file
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		return nil // new file already exists (or unrelated stat error)
	}
	if err := os.MkdirAll(newDir, 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.Rename(legacyPath, newPath); err != nil {
		return fmt.Errorf("failed to migrate config from %s to %s: %w (you may need to manually move the file and restart)", legacyPath, newPath, err)
	}
	return nil
}

// initGHLogin fetches and caches the authenticated GitHub login ID.
// Exits with an error if not logged in.
func initGHLogin() {
	var err error
	ghClient, err = api.DefaultRESTClient()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to initialize GitHub API client: %w", err))
	}
	var user struct {
		Login string `json:"login"`
	}
	if err := ghClient.Get("user", &user); err != nil {
		cobra.CheckErr(fmt.Errorf("not logged in to GitHub CLI, run 'gh auth login' to authenticate"))
	}
	ghLoginID = user.Login
}

func initializeConfig() {
	initGHLogin()

	homePath, err := os.UserHomeDir()
	cobra.CheckErr(err)

	configDir := filepath.Join(homePath, ".config", "gh-cgu")
	configName := "config"
	configType := "yml"

	viper.AddConfigPath(configDir)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	configFile := filepath.Join(configDir, fmt.Sprintf("%s.%s", configName, configType))

	// TODO: Remove this migration block in the version after next.
	// Migrate config from legacy path (~/.config/gh-cgu.yaml) to new path (~/.config/gh-cgu/config.yml).
	legacyConfigFile := filepath.Join(homePath, ".config", "gh-cgu.yaml")
	cobra.CheckErr(migrateConfigIfNeeded(legacyConfigFile, configDir, configFile))

	// if config not found, try to pull from Gist first
	if err := viper.ReadInConfig(); err != nil {
		os.MkdirAll(configDir, 0700)
		fmt.Fprintln(os.Stderr, "! No local config found. Syncing profiles from Gist...")
		pullFromGist(configFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintln(os.Stderr, "! No Gist found. Starting with an empty profile list.")
			// WriteConfigAs はファイルを書くが ConfigFileUsed() を設定しない。
			// SetConfigFile を呼ばないと以後の WriteConfig() が失敗するため明示的に登録する。
			cobra.CheckErr(viper.WriteConfigAs(configFile))
			viper.SetConfigFile(configFile)
		}
	} else {
		// Config exists: sync with Gist (push if local is dirty, pull if Gist is newer)
		syncAtStartup(configFile)
	}
}
