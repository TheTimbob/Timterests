package server_test

import (
	"testing"
)

// isolateWorkingDir moves the test into an empty directory so anything resolved
// relative to the working directory cannot pick up the developer's real files.
func isolateWorkingDir(t *testing.T) {
	t.Helper()

	t.Chdir(t.TempDir())
}
