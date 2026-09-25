package index

import (
	"os"
	"testing"
	"time"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

func TestTokenizer(t *testing.T) {
	text := "git commit -m 'feat: update_user_profile and HandleError in /usr/local/bin'"
	tokens := Tokenize(text)

	expectedSubset := []string{"git", "commit", "feat", "update", "user", "profile", "handle", "error", "usr", "local", "bin"}
	tokenMap := make(map[string]bool)
	for _, tok := range tokens {
		tokenMap[tok] = true
	}

	for _, exp := range expectedSubset {
		if !tokenMap[exp] {
			t.Errorf("Expected token %q not found in tokenizer output: %v", exp, tokens)
		}
	}
}

func TestBM25SearchAndRanking(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "eidetic_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := OpenStorage(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	engine, err := NewEngine(storage)
	if err != nil {
		t.Fatal(err)
	}

	docs := []document.Document{
		document.NewDocument(document.SourceShell, "docker run -d -p 8080:80 nginx", "docker run", time.Now()),
		document.NewDocument(document.SourceShell, "kubectl get pods -n kube-system", "kubectl pods", time.Now()),
		document.NewDocument(document.SourceClipboard, "func CalculateECDSA(k *big.Int) Point", "ECDSA snippet", time.Now()),
		document.NewDocument(document.SourceNote, "Meeting notes regarding docker deployment and container limits", "Notes", time.Now()),
	}

	if err := engine.IndexBatch(docs); err != nil {
		t.Fatalf("IndexBatch failed: %v", err)
	}

	// 1. Search for "docker" - should return 2 documents, with exact command ranked highest
	results := engine.Search("docker", SearchFilter{})
	if len(results) < 2 {
		t.Fatalf("Expected at least 2 results for 'docker', got %d", len(results))
	}
	if results[0].Doc.Source != document.SourceShell {
		t.Errorf("Expected top result to be shell command, got %s", results[0].Doc.Source)
	}

	// 2. Fuzzy search for typo "dokcer"
	fuzzyResults := engine.Search("dokcer", SearchFilter{})
	if len(fuzzyResults) == 0 {
		t.Errorf("Expected fuzzy match for 'dokcer', got 0 results")
	}

	// 3. Source filter: only clipboard
	clipResults := engine.Search("CalculateECDSA", SearchFilter{Source: document.SourceClipboard})
	if len(clipResults) != 1 || clipResults[0].Doc.Source != document.SourceClipboard {
		t.Errorf("Expected 1 clipboard result, got %d", len(clipResults))
	}
}

func TestStoragePersistenceAndReload(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "eidetic_persist_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Step 1: Open storage and index a doc
	storage1, err := OpenStorage(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	engine1, err := NewEngine(storage1)
	if err != nil {
		t.Fatal(err)
	}

	doc := document.NewDocument(document.SourceShell, "systemctl status NetworkManager", "systemctl status", time.Now())
	if err := engine1.Index(doc); err != nil {
		t.Fatal(err)
	}
	storage1.Close()

	// Step 2: Reopen storage in fresh engine and verify doc is reloaded
	storage2, err := OpenStorage(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	defer storage2.Close()

	engine2, err := NewEngine(storage2)
	if err != nil {
		t.Fatal(err)
	}

	if engine2.Count() != 1 {
		t.Fatalf("Expected 1 doc reloaded, got %d", engine2.Count())
	}

	results := engine2.Search("NetworkManager", SearchFilter{})
	if len(results) != 1 || results[0].Doc.Content != doc.Content {
		t.Fatalf("Expected to find reloaded doc, got %v", results)
	}
}
