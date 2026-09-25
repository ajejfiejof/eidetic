package collectors

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ajejfiejof/eidetic/pkg/document"
)

// IsSensitiveCommand applies heuristics to prevent indexing secrets.
func IsSensitiveCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	sensitiveTokens := []string{
		"export aws_secret",
		"export secret",
		"export token",
		"export api_key",
		"export password",
		"password=",
		"passwd=",
		"bearer ",
		"private_key",
		"id_rsa",
	}
	for _, tok := range sensitiveTokens {
		if strings.Contains(lower, tok) {
			return true
		}
	}
	return false
}

// CollectShellHistory discovers and parses history across Bash, Zsh, and Fish.
func CollectShellHistory() ([]document.Document, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var allDocs []document.Document

	// 1. Bash history (~/.bash_history)
	bashPath := filepath.Join(home, ".bash_history")
	if docs, err := parseBashHistory(bashPath); err == nil {
		allDocs = append(allDocs, docs...)
	}

	// 2. Zsh history (~/.zsh_history)
	zshPath := filepath.Join(home, ".zsh_history")
	if docs, err := parseZshHistory(zshPath); err == nil {
		allDocs = append(allDocs, docs...)
	}

	// 3. Fish history (~/.local/share/fish/fish_history)
	fishPath := filepath.Join(home, ".local", "share", "fish", "fish_history")
	if docs, err := parseFishHistory(fishPath); err == nil {
		allDocs = append(allDocs, docs...)
	}

	return allDocs, nil
}

func parseBashHistory(path string) ([]document.Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var docs []document.Document
	scanner := bufio.NewScanner(file)
	var lastTime time.Time
	var lastCmd string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check for Bash timestamp (#1710000000)
		if strings.HasPrefix(line, "#") && len(line) > 1 {
			if ts, err := strconv.ParseInt(line[1:], 10, 64); err == nil {
				lastTime = time.Unix(ts, 0)
				continue
			}
		}

		cmd := line
		if len(cmd) < 3 || cmd == lastCmd || IsSensitiveCommand(cmd) {
			continue
		}
		lastCmd = cmd

		ts := lastTime
		if ts.IsZero() {
			ts = time.Now()
		}

		doc := document.NewDocument(document.SourceShell, cmd, cmd, ts)
		doc.Metadata["shell"] = "bash"
		docs = append(docs, doc)
	}

	return docs, scanner.Err()
}

func parseZshHistory(path string) ([]document.Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var docs []document.Document
	scanner := bufio.NewScanner(file)
	var lastCmd string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		cmd := line
		ts := time.Now()

		// Extended Zsh format: ": 1710000000:0;command"
		if strings.HasPrefix(line, ": ") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				timePart := strings.TrimPrefix(parts[0], ": ")
				timeSub := strings.Split(timePart, ":")
				if len(timeSub) > 0 {
					if epoch, err := strconv.ParseInt(timeSub[0], 10, 64); err == nil {
						ts = time.Unix(epoch, 0)
					}
				}
				cmd = parts[1]
			}
		}

		cmd = strings.TrimSpace(cmd)
		if len(cmd) < 3 || cmd == lastCmd || IsSensitiveCommand(cmd) {
			continue
		}
		lastCmd = cmd

		doc := document.NewDocument(document.SourceShell, cmd, cmd, ts)
		doc.Metadata["shell"] = "zsh"
		docs = append(docs, doc)
	}

	return docs, scanner.Err()
}

func parseFishHistory(path string) ([]document.Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var docs []document.Document
	scanner := bufio.NewScanner(file)
	var curCmd string
	var curTime time.Time

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "- cmd: ") {
			curCmd = strings.TrimPrefix(line, "- cmd: ")
		} else if strings.HasPrefix(line, "  when: ") {
			epochStr := strings.TrimPrefix(line, "  when: ")
			if epoch, err := strconv.ParseInt(strings.TrimSpace(epochStr), 10, 64); err == nil {
				curTime = time.Unix(epoch, 0)
			}

			if len(curCmd) >= 3 && !IsSensitiveCommand(curCmd) {
				doc := document.NewDocument(document.SourceShell, curCmd, curCmd, curTime)
				doc.Metadata["shell"] = "fish"
				docs = append(docs, doc)
			}
			curCmd = ""
			curTime = time.Time{}
		}
	}

	return docs, scanner.Err()
}
