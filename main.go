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
			name := args[0]
			email := args[1]

			viper.Set(name+".name", name)
			viper.Set(name+".email", email)
			viper.WriteConfig()

			fmt.Printf("Add profile: %s<%s>", name, email)

			return
		}

		if flagRemove {
			name := args[0]
			email := viper.GetString(name + ".email")

			configMap := viper.AllSettings()
			delete(configMap, name)
			encodedConfig, _ := json.MarshalIndent(configMap, "", " ")
			err := viper.ReadConfig(bytes.NewReader(encodedConfig))
			cobra.CheckErr(err)
			viper.WriteConfig()

			fmt.Printf("Remove profile: %s<%s>", name, email)

			return
		}

		if f, err := os.Stat(".git"); os.IsNotExist(err) || !f.IsDir() {
			fmt.Println("Error: Not in a git directory")
		}

		if len(args) == 1 {
			key := args[0]
			name := viper.GetString(key + ".name")
			email := viper.GetString(key + ".email")

			if name == "" || email == "" {
				err := fmt.Errorf("no such user")
				cobra.CheckErr(err)
			}

			exec.Command("git", "config", "user.name", name).Run()
			exec.Command("git", "config", "user.email", email).Run()

			fmt.Printf("Change Git User: %s<%s>", name, email)

			return
		}

		userNameOut, _ := exec.Command("git", "config", "user.name").Output()
		userEmailOut, _ := exec.Command("git", "config", "user.email").Output()

		// erase \n ...
		userName := strings.Replace(string(userNameOut), "\n", "", 1)
		userEmail := strings.Replace(string(userEmailOut), "\n", "", 1)

		fmt.Printf("Current Git User: %s<%s>", userName, userEmail)
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
