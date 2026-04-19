package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var useCmd = &cobra.Command{
	Use:   "use <key>",
	Short: "Switch the git user of the current repo to a saved profile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		switchToProfile(viperInstance(), args[0])
	},
}

// checkGitDirectory checks if the current directory is a git repository
func checkGitDirectory() error {
	if f, err := os.Stat(".git"); os.IsNotExist(err) || !f.IsDir() {
		return fmt.Errorf("not a git repository")
	}
	return nil
}

// switchToProfile switches to the specified profile
func switchToProfile(v *viper.Viper, key string) {
	if err := checkGitDirectory(); err != nil {
		cobra.CheckErr(err)
	}

	if !v.IsSet(key) {
		cobra.CheckErr(fmt.Errorf("profile %q not found, run 'gh cgu list' to see available profiles", key))
	}

	name := v.GetString(key + ".name")
	email := v.GetString(key + ".email")
	if name == "" || email == "" {
		cobra.CheckErr(fmt.Errorf("profile %q is corrupted: missing name or email", key))
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

	userName := strings.TrimSpace(string(userNameOut))
	userEmail := strings.TrimSpace(string(userEmailOut))

	fmt.Printf("%s <%s>\n", userName, userEmail)
}
