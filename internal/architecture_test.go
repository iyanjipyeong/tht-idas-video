package internal_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

type importRule struct {
	prefix           string
	forbiddenImports []string
}

func TestCleanArchitectureImportBoundaries(t *testing.T) {
	rules := []importRule{
		{
			prefix: "entity/",
			forbiddenImports: []string{
				"idas-video/internal/usecase",
				"idas-video/internal/adapter",
				"idas-video/internal/infrastructure",
				"net/http",
				"database/sql",
				"github.com/jackc/pgx",
			},
		},
		{
			prefix: "usecase/inbound/",
			forbiddenImports: []string{
				"idas-video/internal/usecase/outbound",
				"idas-video/internal/adapter",
				"idas-video/internal/infrastructure",
				"net/http",
				"database/sql",
				"github.com/jackc/pgx",
			},
		},
		{
			prefix: "usecase/outbound/",
			forbiddenImports: []string{
				"idas-video/internal/usecase/inbound",
				"idas-video/internal/adapter",
				"idas-video/internal/infrastructure",
				"net/http",
				"database/sql",
				"github.com/jackc/pgx",
			},
		},
		{
			prefix: "usecase/",
			forbiddenImports: []string{
				"idas-video/internal/adapter",
				"idas-video/internal/infrastructure",
				"net/http",
				"database/sql",
				"github.com/jackc/pgx",
			},
		},
		{
			prefix: "adapter/inbound/",
			forbiddenImports: []string{
				"idas-video/internal/infrastructure",
				"idas-video/internal/adapter/outbound",
			},
		},
		{
			prefix: "infrastructure/",
			forbiddenImports: []string{
				"idas-video/internal/adapter/inbound",
			},
		},
		{
			prefix: "adapter/outbound/",
			forbiddenImports: []string{
				"idas-video/internal/adapter/inbound",
			},
		},
	}

	files := []string{}
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no Go files found for architecture test")
	}

	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		normalized := filepath.ToSlash(strings.TrimPrefix(file, "./"))
		for _, rule := range rules {
			if !strings.HasPrefix(normalized, rule.prefix) {
				continue
			}

			parsed, parseErr := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
			if parseErr != nil {
				t.Fatalf("parse %s: %v", file, parseErr)
			}

			for _, imported := range parsed.Imports {
				path := strings.Trim(imported.Path.Value, `"`)
				for _, forbidden := range rule.forbiddenImports {
					if path == forbidden || strings.HasPrefix(path, forbidden+"/") {
						t.Fatalf("%s imports forbidden package %s", normalized, path)
					}
				}
			}
		}
	}
}
