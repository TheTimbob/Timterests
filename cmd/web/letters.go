package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"timterests/internal/models"
	"timterests/internal/storage"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
	"golang.org/x/crypto/bcrypt"
)

func LettersPageHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage) {
	var component templ.Component

	letters, err := ListLetters(storageInstance)
	if err != nil {
		message := "Failed to fetch letters"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)
		return
	}

	for i := range letters {
		letters[i].Body = storage.RemoveHTMLTags(letters[i].Body)
	}

	component = LettersListPage(letters)

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in LettersPosts: %e", err)
	}
}

func GetLetterHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage, letterID, password string) {

	password_hash, err := GenerateHash(password)
	if err != nil {
		http.Error(w, "Failed to generate password hash", http.StatusInternalServerError)
		return
	}

	letters, err := ListLetters(storageInstance)
	if err != nil {
		http.Error(w, "Failed to fetch letters", http.StatusInternalServerError)
		return
	}

	for _, letter := range letters {
		if letter.ID == letterID {

			if !ValidatePassword(letter.Password, password_hash) {
				http.Error(w, "Invalid password", http.StatusUnauthorized)
				return
			}

			component := LetterPage(letter)
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Fatalf("Error rendering in GetLettereByIDHandler: %e", err)
			}
		}
	}

}

func ListLetters(storageInstance models.Storage) ([]models.Letter, error) {
	var letters []models.Letter

	// Get all letters from the storage
	prefix := "letters/"
	letterFiles, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range letterFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		letter, err := GetLetter(key, id, storageInstance)
		if err != nil {
			return nil, err
		}

		letters = append(letters, *letter)
	}

	return letters, nil
}

func GetLetter(key string, id int, storageInstance models.Storage) (*models.Letter, error) {
	var letter models.Letter
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	// Retrieve file content
	file, err := storage.GetFile(key, localFilePath, storageInstance)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
		return nil, err
	}

	if err := storage.DecodeFile(file, &letter); err != nil {
		log.Fatalf("Failed to decode file: %v", err)
		return nil, err
	}

	body, err := storage.BodyToHTML(letter.Body)
	if err != nil {
		log.Fatalf("Failed to parse the body text into HTML: %v", err)
		return nil, err
	}

	letter.Body = body
	letter.ID = strconv.Itoa(id)
	return &letter, nil
}

func ValidatePassword(letter_password, password_hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(letter_password))

	if err != nil {
		log.Printf("Password comparison failed: %v", err)
		return false
	}
	return true
}

func GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error generating password hash: %v", err)
		return "", err
	}
	return string(hash), nil
}
