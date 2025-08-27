package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagAdd    bool
	flagRemove bool
)

type Profile interface {
	name() string
	email() string
}

var rootCmd = &cobra.Command{
	Use: "gh cgu",
	Args: func(cmd *cobra.Command, args []string) error {
		if flagAdd {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}

			return nil
		}

		if flagRemove {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			return nil
		}

		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if flagAdd {
			addProfile(args[0], args[1])
			return
		}

		if flagRemove {
			removeProfile(args[0])
			return
		}

		if f, err := os.Stat(".git"); os.IsNotExist(err) || !f.IsDir() {
			fmt.Println("Error: Not in a git directory")
		}

		if len(args) == 1 {
			switchToProfile(args[0])
			return
		}

		showCurrentUser()
	},
}

func main() {
	cobra.OnInitialize(initializeConfig)

	rootCmd.PersistentFlags().BoolVar(&flagAdd, "add", false, "Add profile")
	rootCmd.PersistentFlags().BoolVar(&flagRemove, "remove", false, "Remove profile")
	rootCmd.MarkFlagsMutuallyExclusive("add", "remove")

	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

// addProfile adds a new profile with the given name and email
func addProfile(name, email string) {
	viper.Set(name+".name", name)
	viper.Set(name+".email", email)
	err := viper.WriteConfig()
	if err != nil {
		cobra.CheckErr(err)
	}

	fmt.Printf("Add profile: %s<%s>\n", name, email)
}

// removeProfile removes a profile by name
func removeProfile(name string) {
	email := viper.GetString(name + ".email")

	// Check if profile exists
	if email == "" {
		err := fmt.Errorf("profile '%s' not found", name)
		cobra.CheckErr(err)
	}

	// Get all settings and remove the profile
	configMap := viper.AllSettings()
	delete(configMap, name)

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

	fmt.Printf("Remove profile: %s<%s>\n", name, email)
}

// switchToProfile switches to the specified profile
func switchToProfile(profileName string) {
	name := viper.GetString(profileName + ".name")
	email := viper.GetString(profileName + ".email")

	if name == "" || email == "" {
		err := fmt.Errorf("profile '%s' not found", profileName)
		cobra.CheckErr(err)
	}

	err := exec.Command("git", "config", "user.name", name).Run()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to set git user.name: %w", err))
	}

	err = exec.Command("git", "config", "user.email", email).Run()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to set git user.email: %w", err))
	}

	fmt.Printf("Change Git User: %s<%s>\n", name, email)
}

// showCurrentUser displays the current git user
func showCurrentUser() {
	userNameOut, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to get git user.name: %w", err))
	}

	userEmailOut, err := exec.Command("git", "config", "user.email").Output()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("failed to get git user.email: %w", err))
	}

	// trim newlines
	userName := strings.TrimSpace(string(userNameOut))
	userEmail := strings.TrimSpace(string(userEmailOut))

	fmt.Printf("Current Git User: %s<%s>\n", userName, userEmail)
}

func initializeConfig() {
	homePath, err := os.UserHomeDir()
	cobra.CheckErr(err)

	configPath := filepath.Join(homePath, ".config")
	configName := "gh-cgu"
	configType := "yaml"

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	// if config not found
	if err := viper.ReadInConfig(); err != nil {
		os.MkdirAll(configPath, 0700)
		viper.WriteConfigAs(filepath.Join(configPath, fmt.Sprintf("%s.%s", configName, configType)))
	}
}
