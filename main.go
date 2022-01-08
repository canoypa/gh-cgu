package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	userNameOut, _ := exec.Command("git", "config", "user.name").Output()
	userEmailOut, _ := exec.Command("git", "config", "user.email").Output()

	// erase \n ...
	userName := strings.Replace(string(userNameOut), "\n", "", 1)
	userEmail := strings.Replace(string(userEmailOut), "\n", "", 1)

	fmt.Printf("Current Git User: %s<%s>", userName, userEmail)
}
