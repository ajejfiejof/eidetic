package index

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

// Storage handles durable persistence of documents in an append-only JSONL file.
type Storage struct {
	mu       sync.RWMutex
	filePath string
	file     *os.File
}

// OpenStorage initializes or connects to the append-only document archive.
func OpenStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dataDir, "documents.jsonl")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &Storage{
		filePath: filePath,
		file:     file,
	}, nil
}

// Append writes a document to disk.
func (s *Storage) Append(doc document.Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	if _, err := s.file.Write(append(data, '\n')); err != nil {
		return err
	}
	return s.file.Sync()
}

// AppendBatch writes multiple documents efficiently.
func (s *Storage) AppendBatch(docs []document.Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writer := bufio.NewWriter(s.file)
	for _, doc := range docs {
		data, err := json.Marshal(doc)
		if err != nil {
			continue
		}
		if _, err := writer.Write(append(data, '\n')); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return s.file.Sync()
}

// LoadAll reads all documents from the store, deduplicating by ID.
func (s *Storage) LoadAll() ([]document.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var docs []document.Document
	seen := make(map[string]struct{})

	scanner := bufio.NewScanner(file)
	// Support long lines (up to 4MB)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var doc document.Document
		if err := json.Unmarshal(line, &doc); err != nil {
			continue
		}

		if _, exists := seen[doc.ID]; !exists {
			seen[doc.ID] = struct{}{}
			docs = append(docs, doc)
		}
	}

	return docs, scanner.Err()
}

// Close closes the underlying file descriptor.
func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
