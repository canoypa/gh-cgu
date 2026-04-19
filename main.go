package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	gistDescription = "gh-cgu profiles"
)

var (
	flagKey       string
	flagEditName  string
	flagEditEmail string
	flagEditKey   string
	ghLoginID     string
	ghClient      *api.RESTClient
)

// toKey converts a display name to a profile key.
// Spaces and underscores are replaced with hyphens; leading/trailing hyphens are trimmed.
func toKey(s string) string {
	s = strings.NewReplacer(" ", "-", "_", "-").Replace(s)
	s = strings.Trim(s, "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}

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
	bin, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(bin, "_sync-gist")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = newSysProcAttr()
	cmd.Start()
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
	if gistID == "" {
		ghClient.Post("gists", bytes.NewReader(payloadBytes), &result)
	} else {
		ghClient.Patch(fmt.Sprintf("gists/%s", gistID), bytes.NewReader(payloadBytes), &result)
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

	os.WriteFile(configFile, []byte(f.Content), 0600)
}

var rootCmd = &cobra.Command{
	Use:   "cgu",
	Short: "Manage and switch git user profiles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		showCurrentUser()
	},
}

var useCmd = &cobra.Command{
	Use:   "use <key>",
	Short: "Switch the git user of the current repo to a saved profile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		switchToProfile(args[0])
	},
}

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
		addProfile(args[0], args[1], key)
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <key>",
	Short: "Delete a saved profile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		removeProfile(args[0])
	},
}

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
		editProfile(args[0], flagEditName, flagEditEmail, flagEditKey, cmd.Flags().Changed("name"), cmd.Flags().Changed("email"), cmd.Flags().Changed("key"))
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved profiles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		listProfiles()
	},
}

func main() {
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

	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

// editProfile updates an existing profile's name, email, and/or key
func editProfile(key, name, email, newKey string, changeName, changeEmail, changeKey bool) {
	currentName := viper.GetString(key + ".name")
	currentEmail := viper.GetString(key + ".email")

	if currentName == "" || currentEmail == "" {
		cobra.CheckErr(fmt.Errorf("profile %q not found, run 'gh cgu list' to see available profiles", key))
	}

	if !changeName {
		name = currentName
	}
	if !changeEmail {
		email = currentEmail
	}

	if changeKey {
		// Remove old key and write under new key
		configMap := viper.AllSettings()
		delete(configMap, strings.ToLower(key))
		encodedConfig, err := json.MarshalIndent(configMap, "", " ")
		if err != nil {
			cobra.CheckErr(err)
		}
		if err := viper.ReadConfig(bytes.NewReader(encodedConfig)); err != nil {
			cobra.CheckErr(err)
		}
		key = newKey
	}

	viper.Set(key+".name", name)
	viper.Set(key+".email", email)
	if err := viper.WriteConfig(); err != nil {
		cobra.CheckErr(err)
	}

	fmt.Printf("✓ Updated profile %q (%s <%s>)\n", key, name, email)
	syncToGist()
}

// addProfile adds a new profile with the given display name, email, and key
func addProfile(name, email, key string) {
	viper.Set(key+".name", name)
	viper.Set(key+".email", email)
	err := viper.WriteConfig()
	if err != nil {
		cobra.CheckErr(err)
	}

	fmt.Printf("✓ Added profile %q (%s <%s>)\n", key, name, email)
	syncToGist()
}

// removeProfile removes a profile by key
func removeProfile(key string) {
	email := viper.GetString(key + ".email")

	// Check if profile exists
	if email == "" {
		cobra.CheckErr(fmt.Errorf("profile %q not found, run 'gh cgu list' to see available profiles", key))
	}

	// Get all settings and remove the profile
	// viper lowercases all keys, so match accordingly
	configMap := viper.AllSettings()
	delete(configMap, strings.ToLower(key))

	// Re-encode and reload configuration
	encodedConfig, err := json.MarshalIndent(configMap, "", " ")
	if err != nil {
		cobra.CheckErr(err)
	}

	err = viper.ReadConfig(bytes.NewReader(encodedConfig))
	if err != nil {
		cobra.CheckErr(err)
	}

	err = viper.WriteConfig()
	if err != nil {
		cobra.CheckErr(err)
	}

	fmt.Printf("✓ Removed profile %q\n", key)
	syncToGist()
}

// listProfiles prints all registered profiles
func listProfiles() {
	profiles := viper.AllSettings()
	if len(profiles) == 0 {
		fmt.Println("No profiles found. Add one with: gh cgu add <name> <email>")
		return
	}

	// collect rows and compute column widths
	type row struct{ key, name, email string }
	rows := make([]row, 0, len(profiles))
	keyW, nameW, emailW := len("KEY"), len("NAME"), len("EMAIL")
	for k, v := range profiles {
		entry, ok := v.(map[string]interface{})
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
	for _, r := range rows {
		fmt.Printf("%-*s  %-*s  %-*s\n", keyW, r.key, nameW, r.name, emailW, r.email)
	}
}

// checkGitDirectory checks if the current directory is a git repository
func checkGitDirectory() error {
	if f, err := os.Stat(".git"); os.IsNotExist(err) || !f.IsDir() {
		return fmt.Errorf("not a git repository")
	}
	return nil
}

// switchToProfile switches to the specified profile
func switchToProfile(key string) {
	if err := checkGitDirectory(); err != nil {
		cobra.CheckErr(err)
	}

	name := viper.GetString(key + ".name")
	email := viper.GetString(key + ".email")

	if name == "" || email == "" {
		cobra.CheckErr(fmt.Errorf("profile %q not found, run 'gh cgu list' to see available profiles", key))
	}

	err := exec.Command("git", "config", "user.name", name).Run()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to set git user.name: %w", err))
	}

	err = exec.Command("git", "config", "user.email", email).Run()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to set git user.email: %w", err))
	}

	fmt.Printf("✓ Switched to %q — %s <%s>\n", key, name, email)
}

// showCurrentUser displays the current git user
func showCurrentUser() {
	if err := checkGitDirectory(); err != nil {
		cobra.CheckErr(err)
	}

	userNameOut, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("git user is not set, run 'gh cgu use <key>' to configure one"))
	}

	userEmailOut, err := exec.Command("git", "config", "user.email").Output()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("git user is not set, run 'gh cgu use <key>' to configure one"))
	}

	// trim newlines
	userName := strings.TrimSpace(string(userNameOut))
	userEmail := strings.TrimSpace(string(userEmailOut))

	fmt.Printf("%s <%s>\n", userName, userEmail)
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
		}
	}
}
