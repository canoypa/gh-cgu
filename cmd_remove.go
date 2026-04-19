package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var removeCmd = &cobra.Command{
	Use:   "remove <key>",
	Short: "Delete a saved profile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		removeProfile(viperInstance(), args[0])
	},
}

// removeProfile removes a profile by key
func removeProfile(v *viper.Viper, key string) {
	email := v.GetString(key + ".email")

	if email == "" {
		cobra.CheckErr(fmt.Errorf("profile %q not found, run 'gh cgu list' to see available profiles", key))
	}

	// viper lowercases all keys, so match accordingly
	configMap := v.AllSettings()
	delete(configMap, strings.ToLower(key))

	encodedConfig, err := json.MarshalIndent(configMap, "", " ")
	if err != nil {
		cobra.CheckErr(err)
	}

	err = v.ReadConfig(bytes.NewReader(encodedConfig))
	if err != nil {
		cobra.CheckErr(err)
	}

	err = v.WriteConfig()
	if err != nil {
		cobra.CheckErr(err)
	}

	fmt.Printf("✓ Removed profile %q\n", key)
	syncToGist()
}
