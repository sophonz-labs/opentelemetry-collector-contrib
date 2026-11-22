package migrationmanager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The migration runner splits each file on statement terminators and sends
// every chunk to ClickHouse. A chunk that contains only comments arrives as an
// empty query and fails the whole migration, leaving the version dirty.
//
// Two ways to write one by accident, both of which have happened in this repo:
// a terminator character inside a comment, and a note left after the last
// statement. This test refuses both, because the failure only shows up when the
// migrator runs against a real database.
func TestMigrationsHaveNoCommentOnlyChunks(t *testing.T) {
	root := "migrators"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		body := string(raw)

		for i, line := range strings.Split(body, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "--") && strings.Contains(trimmed, ";") {
				t.Errorf("%s:%d: comment contains a statement terminator, which splits the statement", path, i+1)
			}
		}

		if last := strings.LastIndex(body, ";"); last == -1 {
			if strings.TrimSpace(body) != "" {
				t.Errorf("%s: no statement terminator", path)
			}
		} else if strings.TrimSpace(body[last+1:]) != "" {
			t.Errorf("%s: content after the last statement becomes a comment-only chunk", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk migrations: %v", err)
	}
}
