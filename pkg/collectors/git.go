package collectors

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

// CollectGitCommits searches for git repositories in rootDirs and indexes recent commits.
func CollectGitCommits(rootDirs []string, maxCommitsPerRepo int) ([]document.Document, error) {
	if maxCommitsPerRepo <= 0 {
		maxCommitsPerRepo = 30
	}

	var docs []document.Document
	seenRepos := make(map[string]struct{})

	for _, root := range rootDirs {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() && d.Name() == ".git" {
				repoDir := filepath.Dir(path)
				if _, exists := seenRepos[repoDir]; exists {
					return filepath.SkipDir
				}
				seenRepos[repoDir] = struct{}{}

				repoDocs := indexRepoCommits(repoDir, maxCommitsPerRepo)
				docs = append(docs, repoDocs...)
				return filepath.SkipDir
			}

			// Don't recurse more than 3 levels looking for repos
			rel, _ := filepath.Rel(root, path)
			if strings.Count(rel, string(os.PathSeparator)) > 3 {
				return filepath.SkipDir
			}

			return nil
		})
	}

	return docs, nil
}

func indexRepoCommits(repoDir string, maxCommits int) []document.Document {
	repoName := filepath.Base(repoDir)

	// Format: %H%x00%an%x00%ct%x00%s%x00%b%x1e
	// hash, author, unix_timestamp, subject, body
	format := "%H%x00%an%x00%ct%x00%s%x00%b%x1e"
	cmd := exec.Command("git", "-C", repoDir, "log", "-n", strconv.Itoa(maxCommits), "--format="+format)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return nil
	}

	var docs []document.Document
	records := bytes.Split(out, []byte{0x1e})

	for _, rec := range records {
		rec = bytes.TrimSpace(rec)
		if len(rec) == 0 {
			continue
		}

		fields := bytes.Split(rec, []byte{0x00})
		if len(fields) < 4 {
			continue
		}

		hash := string(fields[0])
		author := string(fields[1])
		epochStr := string(fields[2])
		subject := string(fields[3])
		body := ""
		if len(fields) >= 5 {
			body = strings.TrimSpace(string(fields[4]))
		}

		epoch, _ := strconv.ParseInt(epochStr, 10, 64)
		ts := time.Unix(epoch, 0)

		shortHash := hash
		if len(shortHash) > 8 {
			shortHash = shortHash[:8]
		}

		title := fmt.Sprintf("[%s] %s: %s", repoName, shortHash, subject)
		content := fmt.Sprintf("commit %s\nAuthor: %s\nDate: %s\nRepository: %s\n\n    %s",
			hash, author, ts.Format(time.RFC1123), repoDir, subject)
		if body != "" {
			content += "\n\n    " + strings.ReplaceAll(body, "\n", "\n    ")
		}

		doc := document.NewDocument(document.SourceGit, content, title, ts)
		doc.Metadata["repo"] = repoName
		doc.Metadata["repo_path"] = repoDir
		doc.Metadata["hash"] = hash
		doc.Metadata["author"] = author
		doc.Tags = []string{"git", repoName}

		docs = append(docs, doc)
	}

	return docs
}
