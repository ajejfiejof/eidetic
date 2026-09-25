package collectors

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

// CollectStdin reads piped input from standard input.
func CollectStdin(title string, tags []string) (document.Document, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return document.Document{}, err
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return document.Document{}, errors.New("stdin is empty (no piped input)")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return document.Document{}, err
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return document.Document{}, errors.New("piped input is empty")
	}

	doc := document.NewDocument(document.SourceStdin, content, title, time.Now())
	doc.Tags = tags
	return doc, nil
}

// CollectSnippet creates a document from user CLI parameters.
func CollectSnippet(content, title string, tags []string) document.Document {
	doc := document.NewDocument(document.SourceNote, content, title, time.Now())
	doc.Tags = tags
	return doc
}
