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

// initGHLogin fetches and caches the authenticated GitHub login ID.
// Exits with an error if not logged in.
func initGHLogin() {
	var err error
	ghClient, err = api.DefaultRESTClient()
	if err != nil {
		cobra.CheckErr(err)
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

	configPath := filepath.Join(homePath, ".config")
	configName := "gh-cgu"
	configType := "yaml"

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	configFile := filepath.Join(configPath, fmt.Sprintf("%s.%s", configName, configType))

	// if config not found, try to pull from Gist first
	if err := viper.ReadInConfig(); err != nil {
		os.MkdirAll(configPath, 0700)
		fmt.Fprintln(os.Stderr, "! No local config found. Syncing profiles from Gist...")
		pullFromGist(configFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintln(os.Stderr, "! No Gist found. Starting with an empty profile list.")
			viper.WriteConfigAs(configFile)
			// WriteConfigAs はファイルを書くが ConfigFileUsed() を設定しない。
			// SetConfigFile を呼ばないと以後の WriteConfig() が失敗するため明示的に登録する。
			viper.SetConfigFile(configFile)
		}
	}
}
