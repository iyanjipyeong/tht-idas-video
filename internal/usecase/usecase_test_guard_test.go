package usecase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryUsecaseFileHasMatchingUnitTest(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	testFiles := map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, "_test.go") {
			continue
		}
		testFiles[name] = true
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, "_usecase.go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		expectedTestName := strings.TrimSuffix(name, ".go") + "_test.go"
		if !testFiles[expectedTestName] {
			t.Fatalf("missing unit test file for usecase %s; expected %s", filepath.Base(name), expectedTestName)
		}
	}
}
