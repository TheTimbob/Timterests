package service_test

import (
	"context"
	"path/filepath"
	"testing"
	"timterests/internal/storage"
)

// testSetup initialises a Storage instance pointing at the shared testdata directory.
func testSetup(t *testing.T, ctx context.Context) *storage.Storage {
	t.Helper()
	t.Setenv("USE_S3", "false")

	s, err := storage.NewStorage(ctx)
	if err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	// Point at the existing testdata fixtures used by the web package tests.
	s.BaseDir = filepath.Join(s.BaseDir, "testdata")

	return s
}
