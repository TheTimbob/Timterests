package service

import (
	"context"
	"fmt"
	"strconv"
	"timterests/internal/storage"
)

// metaSetter is satisfied by any type whose pointer embeds *model.Document.
type metaSetter interface {
	SetMeta(id, key string)
}

// getDoc initialises a zero-value T, sets its metadata, fetches and prepares
// the file from storage, and returns a pointer to the result.
func getDoc[T any, PT interface {
	*T
	metaSetter
}](ctx context.Context, s storage.Storage, key string, id int) (*T, error) {
	var doc T

	PT(&doc).SetMeta(strconv.Itoa(id), key)

	err := s.GetPreparedFile(ctx, key, PT(&doc))
	if err != nil {
		return nil, fmt.Errorf("failed to get prepared file: %w", err)
	}

	return &doc, nil
}
