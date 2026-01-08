package web_test

import (
	"context"
	"path/filepath"
	"testing"
	"timterests/internal/storage"
)

func testSetup(t *testing.T, ctx context.Context) *storage.Storage {
	t.Helper()
	t.Setenv("USE_S3", "false")

	s, err := storage.NewStorage(ctx)
	if err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	s.BaseDir = filepath.Join(s.BaseDir, "testdata")

	return s
}
