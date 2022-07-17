package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getConfigPath() string {
	path := os.Getenv("HOME")

	if path == "" && runtime.GOOS == "windows" {
		path = os.Getenv("APPDATA")
	} else {
		path = filepath.Join(path, ".config")
	}

	path = filepath.Join(path, "gh-cgu")

	return path
}

var flagSet bool

var rootCmd = &cobra.Command{
	Use: "gh cgu",
	Run: func(c *cobra.Command, args []string) {

		// show current user
		// e.g. gh cgu
		if len(args) == 0 {
			userNameOut, _ := exec.Command("git", "config", "user.name").Output()
			userEmailOut, _ := exec.Command("git", "config", "user.email").Output()

			// erase \n ...
			userName := strings.Replace(string(userNameOut), "\n", "", 1)
			userEmail := strings.Replace(string(userEmailOut), "\n", "", 1)

			fmt.Printf("Current Git User: %s<%s>", userName, userEmail)

			return
		}

		// set user
		// e.g. gh cgu --set username user@example.com
		if flagSet && len(args) == 2 || len(args) == 3 {
			name := args[0]
			gitUserName := args[0]
			gitUserEmail := args[1]

			if len(args) == 3 {
				gitUserName = args[1]
				gitUserEmail = args[2]
			}

			viper.Set(name+".name", gitUserName)
			viper.Set(name+".email", gitUserEmail)
			viper.WriteConfig()

			fmt.Printf("Add Git User: %s<%s>", gitUserName, gitUserEmail)

			return
		}

		// change user
		// e.g. gh cgu username
		if len(args) == 1 {
			if f, err := os.Stat(".git"); os.IsNotExist(err) || !f.IsDir() {
				fmt.Println("Error: Not in a git directory")
			}

			getName := args[0]

			settingName := viper.GetString(getName + ".name")
			settingEmail := viper.GetString(getName + ".email")

			if settingName == "" || settingEmail == "" {
				fmt.Println("Error: No such user")
				return
			}

			exec.Command("git", "config", "user.name", settingName).Run()
			exec.Command("git", "config", "user.email", settingEmail).Run()

			fmt.Printf("Change Git User: %s<%s>", settingName, settingEmail)

			return
		}

		c.Help()
	},
}

func main() {
	cobra.OnInitialize(func() {
		configName := "config"
		configType := "yaml"
		configPath := getConfigPath()

		viper.SetConfigName(configName)
		viper.SetConfigType(configType)
		viper.AddConfigPath(configPath)

		if err := viper.ReadInConfig(); err != nil {
			os.MkdirAll(configPath, 0700)
			viper.WriteConfigAs(filepath.Join(configPath, fmt.Sprintf("%s.%s", configName, configType)))
		}
	})

	rootCmd.PersistentFlags().BoolVarP(&flagSet, "set", "s", false, "set user")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
