package collectors

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

// ReadClipboard attempts to read the current text from system clipboard (X11 or Wayland).
func ReadClipboard() (string, error) {
	// Try Wayland first
	if path, err := exec.LookPath("wl-paste"); err == nil {
		out, err := exec.Command(path, "--no-newline").Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}

	// Try xclip (X11)
	if path, err := exec.LookPath("xclip"); err == nil {
		out, err := exec.Command(path, "-selection", "clipboard", "-o").Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}

	// Try xsel (X11)
	if path, err := exec.LookPath("xsel"); err == nil {
		out, err := exec.Command(path, "--clipboard", "--output").Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}

	return "", errors.New("no supported clipboard utility found (need wl-paste, xclip, or xsel)")
}

// WatchClipboard monitors system clipboard changes and dispatches new items.
func WatchClipboard(ctx context.Context, interval time.Duration, onNewItem func(document.Document)) {
	if interval < 500*time.Millisecond {
		interval = 1 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastHash string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			text, err := ReadClipboard()
			if err != nil || len(text) < 4 || len(text) > 64*1024 {
				continue
			}

			// Filter out passwords and secrets
			if IsSensitiveCommand(text) {
				continue
			}

			doc := document.NewDocument(document.SourceClipboard, text, "", time.Now())
			if doc.ID == lastHash {
				continue
			}
			lastHash = doc.ID

			onNewItem(doc)
		}
	}
}
