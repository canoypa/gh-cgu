package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func viperInstance() *viper.Viper {
	return viper.GetViper()
}

var (
	flagKey       string
	flagEditName  string
	flagEditEmail string
	flagEditKey   string
)

var rootCmd = &cobra.Command{
	Use:   "cgu",
	Short: "Manage and switch git user profiles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		showCurrentUser()
	},
}

func init() {
	cobra.OnInitialize(initializeConfig)

	addCmd.Flags().StringVar(&flagKey, "key", "", "Profile key used in config (defaults to name with spaces replaced by hyphens)")
	editCmd.Flags().StringVar(&flagEditName, "name", "", "New display name")
	editCmd.Flags().StringVar(&flagEditEmail, "email", "", "New email address")
	editCmd.Flags().StringVar(&flagEditKey, "key", "", "New profile key")

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(useCmd, addCmd, editCmd, removeCmd, listCmd, syncGistCmd)
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  gh {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  gh {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasAvailableSubCommands}}

Use "gh {{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}
