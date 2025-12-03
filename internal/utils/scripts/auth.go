package scripts

import (
	"errors"
	"fmt"
	"timterests/internal/auth"
	"timterests/internal/storage"
)

func CreateUser(firstName, lastName, email, password string) error {
	if email == "" || password == "" {
		return errors.New("email and password cannot be empty")
	}

	if err := storage.InitDB(); err != nil {
		fmt.Println(err)
		return err
	}

	err := auth.CreateUser(firstName, lastName, email, password)
	if err != nil {
		return err
	}
	return nil
}
