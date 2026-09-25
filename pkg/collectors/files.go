package collectors

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

var allowedExtensions = map[string]bool{
	".md":       true,
	".markdown": true,
	".txt":      true,
	".org":      true,
	".sh":       true,
	".bash":     true,
	".zsh":      true,
	".go":       true,
	".py":       true,
	".rs":       true,
	".json":     true,
	".yaml":     true,
	".yml":      true,
	".toml":     true,
	".conf":     true,
	".ini":      true,
}

// CollectFiles crawls user directories and indexes markdown files, notes, and scripts.
func CollectFiles(rootDirs []string, maxDocSizeKB int, ignorePatterns []string) ([]document.Document, error) {
	if maxDocSizeKB <= 0 {
		maxDocSizeKB = 512
	}
	maxBytes := int64(maxDocSizeKB * 1024)

	var docs []document.Document

	for _, root := range rootDirs {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			name := d.Name()
			if d.IsDir() {
				// Skip hidden directories and ignore rules
				if strings.HasPrefix(name, ".") && name != "." {
					return filepath.SkipDir
				}
				for _, ig := range ignorePatterns {
					if strings.EqualFold(name, ig) {
						return filepath.SkipDir
					}
				}
				return nil
			}

			// File checks
			ext := strings.ToLower(filepath.Ext(path))
			if !allowedExtensions[ext] {
				return nil
			}

			info, err := d.Info()
			if err != nil || info.Size() == 0 || info.Size() > maxBytes {
				return nil
			}

			contentBytes, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			// Skip binary files (null byte check)
			if bytes.IndexByte(contentBytes, 0) != -1 {
				return nil
			}

			content := strings.TrimSpace(string(contentBytes))
			if len(content) < 10 {
				return nil
			}

			// Extract title
			title := filepath.Base(path)
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				lineTrimmed := strings.TrimSpace(line)
				if strings.HasPrefix(lineTrimmed, "# ") {
					title = strings.TrimPrefix(lineTrimmed, "# ")
					break
				}
			}

			sourceType := document.SourceFile
			if ext == ".md" || ext == ".txt" || ext == ".org" {
				sourceType = document.SourceNote
			}

			doc := document.NewDocument(sourceType, content, title, info.ModTime())
			doc.Metadata["path"] = path
			doc.Metadata["ext"] = ext
			doc.Metadata["bytes"] = strconv.FormatInt(info.Size(), 10)

			docs = append(docs, doc)
			return nil
		})

		if err != nil {
			continue
		}
	}

	return docs, nil
}
