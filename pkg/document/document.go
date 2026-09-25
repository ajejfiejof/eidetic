package document

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// SourceType represents the origin of an indexed item.
type SourceType string

const (
	SourceShell     SourceType = "shell"
	SourceClipboard SourceType = "clipboard"
	SourceFile      SourceType = "file"
	SourceNote      SourceType = "note"
	SourceStdin     SourceType = "stdin"
	SourceGit       SourceType = "git"
)

// Document represents an indexed knowledge artifact with metadata.
type Document struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Source    SourceType        `json:"source"`
	Title     string            `json:"title"`
	Tags      []string          `json:"tags,omitempty"`
	Pinned    bool              `json:"pinned,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ComputeID calculates a deterministic content-addressable hash.
func ComputeID(source SourceType, content string, extra ...string) string {
	hasher := sha256.New()
	hasher.Write([]byte(string(source)))
	hasher.Write([]byte{0})
	hasher.Write([]byte(strings.TrimSpace(content)))
	for _, e := range extra {
		hasher.Write([]byte{0})
		hasher.Write([]byte(e))
	}
	return hex.EncodeToString(hasher.Sum(nil))[:16]
}

// NewDocument creates a new document with an automatic deterministic ID.
func NewDocument(source SourceType, content string, title string, timestamp time.Time) Document {
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	trimmed := strings.TrimSpace(content)
	if title == "" {
		lines := strings.Split(trimmed, "\n")
		if len(lines) > 0 {
			title = strings.TrimSpace(lines[0])
			if len(title) > 60 {
				title = title[:57] + "..."
			}
		}
	}

	return Document{
		ID:        ComputeID(source, trimmed, title),
		Content:   trimmed,
		Source:    source,
		Title:     title,
		Timestamp: timestamp,
		Metadata:  make(map[string]string),
	}
}
