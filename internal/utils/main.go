package main

import (
	"fmt"
	"os"
	"timterests/internal/utils/scripts"
)

func main() {
	var command string
	var arguments []string
	if len(os.Args) < 2 {
		command = "help"
	} else {
		command = os.Args[1]
		arguments = os.Args[2:]
	}

	switch command {
	case "help":
		fmt.Println("Usage: go run main.go <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  create-user <firstName> <lastName> <email> <password>  Create a new user")
		return
	case "create-user":
		if len(arguments) != 4 {
			fmt.Println("Usage: create-user <firstName> <lastName> <email> <password>")
			return
		}
		err := scripts.CreateUser(arguments[0], arguments[1], arguments[2], arguments[3])
		if err != nil {
			panic(err)
		}
		return
	default:
		fmt.Println("Unknown command")
		return
	}
}
