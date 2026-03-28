package service

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"timterests/internal/model"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// ListLetters retrieves all letters from storage, optionally filtering by tag.
// Pass tag="" or tag="all" to retrieve all letters.
func ListLetters(ctx context.Context, s storage.Storage, tag string) ([]model.Letter, error) {
	prefix := "letters/"

	letterFiles, err := s.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	letters := make([]model.Letter, 0, len(letterFiles))

	for id, obj := range letterFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		letter, err := GetLetter(ctx, s, key, id)
		if err != nil {
			return nil, err
		}

		if tag == "all" || tag == "" || slices.Contains(letter.Tags, tag) {
			letters = append(letters, *letter)
		}
	}

	sort.Slice(letters, func(i, j int) bool {
		return letters[i].Date > letters[j].Date
	})

	return letters, nil
}

// GetLetter retrieves a single letter by its storage key and numeric ID.
func GetLetter(ctx context.Context, s storage.Storage, key string, id int) (*model.Letter, error) {
	return getDoc[model.Letter](ctx, s, key, id)
}
