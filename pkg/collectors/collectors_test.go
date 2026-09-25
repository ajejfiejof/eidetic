package collectors

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

func TestParseBashHistory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bash_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	histFile := filepath.Join(tempDir, ".bash_history")
	histContent := `#1710000000
git status
#1710000010
docker compose up -d
export AWS_SECRET_ACCESS_KEY=supersecretpassword
#1710000020
ls -la
`
	if err := os.WriteFile(histFile, []byte(histContent), 0644); err != nil {
		t.Fatal(err)
	}

	docs, err := parseBashHistory(histFile)
	if err != nil {
		t.Fatalf("parseBashHistory failed: %v", err)
	}

	// Should contain "git status", "docker compose up -d", and "ls -la"
	// Should NOT contain the secret!
	for _, doc := range docs {
		if IsSensitiveCommand(doc.Content) {
			t.Errorf("Sensitive command leaked into index: %s", doc.Content)
		}
	}

	if len(docs) != 3 {
		t.Fatalf("Expected 3 parsed commands, got %d", len(docs))
	}
}

func TestCollectFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "files_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	notePath := filepath.Join(tempDir, "sample.md")
	noteContent := `# Project Architecture
This is a test document explaining the system.
`
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Fatal(err)
	}

	docs, err := CollectFiles([]string{tempDir}, 512, []string{".git"})
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 collected file, got %d", len(docs))
	}

	if docs[0].Title != "Project Architecture" {
		t.Errorf("Expected title 'Project Architecture', got %q", docs[0].Title)
	}
	if docs[0].Source != document.SourceNote {
		t.Errorf("Expected SourceNote, got %s", docs[0].Source)
	}
}
