package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest profiles from Gist",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			cobra.CheckErr(fmt.Errorf("config file not initialized"))
		}

		m := readSyncMeta()
		if err := pullConfigFromGist(configFile, &m); err != nil {
			fmt.Fprintln(os.Stderr, "! Could not reach Gist (using local config)")
			return
		}
		writeSyncMeta(m)

		if err := viper.ReadInConfig(); err != nil {
			cobra.CheckErr(fmt.Errorf("failed to reload config: %w", err))
		}
		fmt.Println("✓ Pulled latest profiles from Gist")
	},
}
