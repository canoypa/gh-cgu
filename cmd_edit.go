package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var editCmd = &cobra.Command{
	Use:   "edit <key>",
	Short: "Edit an existing profile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("email") && !cmd.Flags().Changed("key") {
			cmd.Help()
			return
		}
		editProfile(viperInstance(), args[0], flagEditName, flagEditEmail, flagEditKey, cmd.Flags().Changed("name"), cmd.Flags().Changed("email"), cmd.Flags().Changed("key"))
	},
}

// editProfile updates an existing profile's name, email, and/or key
func editProfile(v *viper.Viper, key, name, email, newKey string, changeName, changeEmail, changeKey bool) {
	if !v.IsSet(key) {
		cobra.CheckErr(fmt.Errorf("profile %q not found, run 'gh cgu list' to see available profiles", key))
	}

	currentName := v.GetString(key + ".name")
	currentEmail := v.GetString(key + ".email")
	if currentName == "" || currentEmail == "" {
		cobra.CheckErr(fmt.Errorf("profile %q is corrupted: missing name or email", key))
	}

	if !changeName {
		name = currentName
	}
	if !changeEmail {
		email = currentEmail
	}

	if changeKey {
		if err := validateKey(newKey); err != nil {
			cobra.CheckErr(err)
		}
		configMap := v.AllSettings()
		delete(configMap, strings.ToLower(key))
		configMap[strings.ToLower(newKey)] = map[string]interface{}{
			"name":  name,
			"email": email,
		}
		encodedConfig, err := json.MarshalIndent(configMap, "", " ")
		if err != nil {
			cobra.CheckErr(err)
		}
		if err := v.ReadConfig(bytes.NewReader(encodedConfig)); err != nil {
			cobra.CheckErr(err)
		}
		key = newKey
	} else {
		configMap := v.AllSettings()
		keyLower := strings.ToLower(key)
		entry, ok := configMap[keyLower].(map[string]interface{})
		if !ok {
			entry = make(map[string]interface{})
		}
		entry["name"] = name
		entry["email"] = email
		configMap[keyLower] = entry
		encodedConfig, err := json.MarshalIndent(configMap, "", " ")
		if err != nil {
			cobra.CheckErr(err)
		}
		if err := v.ReadConfig(bytes.NewReader(encodedConfig)); err != nil {
			cobra.CheckErr(err)
		}
	}

	if err := v.WriteConfig(); err != nil {
		cobra.CheckErr(err)
	}

	fmt.Printf("✓ Updated profile %q (%s <%s>)\n", key, name, email)
	syncToGist()
}
