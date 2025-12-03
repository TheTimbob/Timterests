package main

import (
	"fmt"
	"os"
	"timterests/internal/storage"
	"timterests/internal/utils/scripts"
)

func main() {
	var command string
	var arguements []string
	if len(os.Args) < 2 {
		command = "help"
	} else {
		command = os.Args[1]
		arguements = os.Args[2:]
	}

	switch command {
	case "help":
		fmt.Println("Usage: go run main.go <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  create-user <firstName> <lastName> <email> <password>  Create a new user")
		return
	case "create-user":
		err := storage.InitDB()
		if err != nil {
			panic(err)
		}
		err = scripts.CreateUser(arguements[0], arguements[1], arguements[2], arguements[3])
		if err != nil {
			panic(err)
		}
		return
	default:
		fmt.Println("Unknown command")
		return
	}
}
