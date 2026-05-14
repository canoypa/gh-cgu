package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addCmd = &cobra.Command{
	Use:   "add <name> <email>",
	Short: "Save a new git user profile",
	Long: `Save a new git user profile.

The profile key is derived from <name> by replacing spaces with hyphens.
Use --key to set a different key explicitly.`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Help()
			return
		}
		key := flagKey
		if key == "" {
			key = toKey(args[0])
			if key == "" {
				cobra.CheckErr(fmt.Errorf("could not derive key from %q. Use --key to specify one", args[0]))
			}
		}
		addProfile(viperInstance(), args[0], args[1], key)
	},
}

// toKey converts a display name to a profile key.
// Spaces, underscores and dots are replaced with hyphens; leading/trailing hyphens are trimmed.
func toKey(s string) string {
	s = strings.NewReplacer(" ", "-", "_", "-", ".", "-").Replace(s)
	s = strings.Trim(s, "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}

// validKeyRe matches safe profile key characters:
// Unicode letters (\p{L}), digits (\p{N}), hyphens and underscores.
// YAML-special characters (.: # @ ! etc.) and whitespace are excluded.
var validKeyRe = regexp.MustCompile(`^[\p{L}\p{N}_\-]+$`)

// validateKey returns an error if the key contains characters that Viper treats as
// path delimiters (currently ".") or characters that would corrupt the YAML structure.
func validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("profile key must not be empty")
	}
	if !validKeyRe.MatchString(key) {
		return fmt.Errorf("profile key %q contains invalid characters: only letters, digits, hyphens and underscores are allowed", key)
	}
	return nil
}

// addProfile adds a new profile with the given display name, email, and key
func addProfile(v *viper.Viper, name, email, key string) {
	if err := validateKey(key); err != nil {
		cobra.CheckErr(err)
	}
	if v.IsSet(key) {
		cobra.CheckErr(fmt.Errorf("profile %q already exists, use 'gh cgu edit %s' to update it", key, key))
	}

	configMap := v.AllSettings()
	if configMap == nil {
		configMap = make(map[string]interface{})
	}
	configMap[strings.ToLower(key)] = map[string]interface{}{
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
	if err := v.WriteConfig(); err != nil {
		cobra.CheckErr(err)
	}

	fmt.Printf("✓ Added profile %q (%s <%s>)\n", key, name, email)
	syncToGist()
}
