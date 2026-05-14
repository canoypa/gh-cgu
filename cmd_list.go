package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved profiles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		listProfiles(viperInstance())
	},
}

// listProfiles prints all registered profiles
func listProfiles(v *viper.Viper) {
	profiles := v.AllSettings()
	if len(profiles) == 0 {
		fmt.Println("No profiles found. Add one with: gh cgu add <name> <email>")
		return
	}

	type row struct{ key, name, email string }
	rows := make([]row, 0, len(profiles))
	keyW, nameW, emailW := len("KEY"), len("NAME"), len("EMAIL")
	for k, val := range profiles {
		entry, ok := val.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		email, _ := entry["email"].(string)
		rows = append(rows, row{k, name, email})
		if len(k) > keyW {
			keyW = len(k)
		}
		if len(name) > nameW {
			nameW = len(name)
		}
		if len(email) > emailW {
			emailW = len(email)
		}
	}

	fmt.Printf("%-*s  %-*s  %-*s\n", keyW, "KEY", nameW, "NAME", emailW, "EMAIL")
	fmt.Printf("%s  %s  %s\n", strings.Repeat("-", keyW), strings.Repeat("-", nameW), strings.Repeat("-", emailW))
	sort.Slice(rows, func(i, j int) bool { return rows[i].key < rows[j].key })
	for _, r := range rows {
		fmt.Printf("%-*s  %-*s  %-*s\n", keyW, r.key, nameW, r.name, emailW, r.email)
	}
}
