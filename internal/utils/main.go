package main

import (
	"fmt"
	"os"
	"path/filepath"
	"timterests/internal/utils/scripts"
)

func getDBPath() (string, error) {

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v", err)
		return "", err
	}
	rootDir := filepath.Dir(filepath.Dir(cwd))
	dbPath := filepath.Join(rootDir, "database", "timterests.db")

	return dbPath, err
}

func main() {
	command := os.Args[1]
	arguements := os.Args[2:]
	switch command {
	case "help":
		fmt.Println("Usage: go run main.go <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  create-user <firstName> <lastName> <email> <password>  Create a new user")
		return
	case "create-user":
		dbPath, err := getDBPath()
		if err != nil {
			panic(err)
		}
		err = scripts.CreateUser(dbPath, arguements[0], arguements[1], arguements[2], arguements[3])
		if err != nil {
			panic(err)
		}
		return
	default:
		fmt.Println("Unknown command")
		return
	}
}
