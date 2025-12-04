package scripts

import (
	"timterests/internal/auth"
	"timterests/internal/storage"
)

func CreateUser(firstName, lastName, email, password string) error {
	if err := storage.InitDB(); err != nil {
		return err
	}

	return auth.CreateUser(firstName, lastName, email, password)
}
