package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var flagSet bool

var rootCmd = &cobra.Command{
	Use: "cgu",
	Run: func(c *cobra.Command, args []string) {
		if f, err := os.Stat(".git"); os.IsNotExist(err) || !f.IsDir() {
			fmt.Println("Error: Not in a git directory")
		} else {

			// show current user
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
			if flagSet && len(args) == 2 {
				name := args[0]
				email := args[1]

				viper.Set(name+".name", name)
				viper.Set(name+".email", email)
				viper.WriteConfig()

				fmt.Printf("Add Git User: %s<%s>", name, email)

				return
			}

			// change user
			if len(args) == 1 {
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
		}

	},
}

func main() {
	cobra.OnInitialize(func() {
		viper.SetConfigFile("config.toml")
		viper.AddConfigPath(".")

		viper.ReadInConfig()
	})

	rootCmd.PersistentFlags().BoolVarP(&flagSet, "set", "s", false, "set user")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
