# Eidetic 🧠

> **Your sovereign cognitive index — instant photographic recall across everything you've ever typed, copied, or read.**

`eidetic` is a single-binary, local-first search engine and memory layer for Linux developers. It automatically captures, indexes, and ranks your shell history, clipboard history, notes, and code snippets into an ultra-fast local index, accessible via an interactive terminal UI or instant CLI.

Zero cloud. Zero telemetry. Zero external database services. Runs on your hardware at microsecond latencies.

---

## ⚡ Features

* **Sub-Millisecond Search**: In-memory inverted index powered by **BM25 relevance scoring**, exact phrase matching, and **Levenshtein fuzzy matching** for typo tolerance.
* **Interactive Split-Screen TUI**: As-you-type fuzzy searching built with Bubbletea & Lipgloss. Live preview pane, relative timestamps, and one-key selection/yank.
* **Multi-Shell History Ingestion**: Automatically discovers and indexes `.bash_history`, `.zsh_history`, and Fish shell history with timestamps and secret sanitization.
* **Clipboard Watcher**: Optional background daemon capturing text copies on both **Wayland** (`wl-paste`) and **X11** (`xclip`/`xsel`).
* **Local Notes & Snippet Crawler**: Recursively indexes markdown notes, scripts, and configuration files from `~/Notes`, `~/.config`, etc.
* **Piped Stdin Capture**: Easily archive command outputs or logs: `cat deploy.log | eidetic add --title "Deployment log"`.
* **Shell Integration (`Ctrl+R`)**: Seamlessly replaces terminal reverse-i-search with Eidetic's rich semantic memory.

---

## 🚀 Quick Start

### 1. Build & Install
```bash
git clone https://github.com/ajejfiejof/eidetic.git
cd eidetic
go build -o bin/eidetic ./cmd/eidetic

# Optional: install to your PATH
sudo cp bin/eidetic /usr/local/bin/
```

### 2. Index Your System
```bash
# Indexes shell history, clipboard, and local notes
eidetic index
```

### 3. Search
```bash
# Launch interactive TUI
eidetic

# Or search from the command line
eidetic search docker
```

---

## ⌨️ Interactive TUI Navigation

Running `eidetic` without arguments opens the full-screen terminal interface:

| Keybinding | Action |
|---|---|
| `Type any text` | Instant as-you-type fuzzy search |
| `↑` / `Ctrl+K` / `Ctrl+P` | Move selection up |
| `↓` / `Ctrl+J` / `Ctrl+N` | Move selection down |
| `Tab` | Cycle source filter (`ALL` → `SHELL` → `CLIP` → `FILE` → `NOTE`) |
| `Enter` | Select and print item (or insert into shell) |
| `Backspace` | Delete search characters |
| `Esc` / `Ctrl+C` | Exit |

---

## 💻 CLI Commands

### Direct Search
```bash
# Search across all sources
eidetic search "git commit"

# Filter specifically by source
eidetic search --source shell "systemctl"
eidetic search --source clip "curl"
```

### Manual Note & Stdin Capture
```bash
# Add a quick snippet
eidetic add "openssl s_client -connect example.com:443 -servername example.com"

# Pipe output from any command
find . -type f -name '*.go' | xargs wc -l | eidetic add
```

### Clipboard Daemon
```bash
# Run background clipboard watcher
eidetic watch
```

### Shell Integration (Upgrade `Ctrl+R`)
Add this to your `~/.bashrc`:
```bash
eval "$(eidetic init bash)"
```
Or for `~/.zshrc`:
```bash
eval "$(eidetic init zsh)"
```

---

## 📊 Performance (Tested on ThinkPad T440p)

* **Startup & Query Latency**: **10 to 100 microseconds** ($0.01\text{ms} - 0.10\text{ms}$) across 2,000+ items.
* **Index Throughput**: **2,380+ items indexed in 176 milliseconds**.
* **Memory Footprint**: $<15\text{MB}$ resident set size.
* **Storage Engine**: Append-only JSONL document store (`~/.local/share/eidetic/documents.jsonl`).

---

## 📁 Repository Structure

```
eidetic/
├── cmd/
│   └── eidetic/
│       └── main.go           # CLI runner & subcommands
├── pkg/
│   ├── config/
│   │   └── config.go         # XDG configuration paths & defaults
│   ├── document/
│   │   └── document.go       # Document model & deterministic ID hashing
│   ├── index/
│   │   ├── engine.go         # Inverted index, BM25 scoring & fuzzy matching
│   │   ├── tokenizer.go      # Code & shell aware tokenizer
│   │   ├── storage.go        # Append-only JSONL persistence
│   │   └── engine_test.go    # Unit tests for search & ranking
│   ├── collectors/
│   │   ├── shell.go          # Bash, Zsh, Fish history parsers
│   │   ├── clipboard.go      # Wayland & X11 clipboard watcher
│   │   ├── files.go          # Notes & markdown crawler
│   │   └── stdin.go          # Piped input ingestion
│   ├── tui/
│   │   ├── model.go          # Bubbletea TUI model & state loop
│   │   └── styles.go         # Lipgloss visual styling & badges
│   └── integration/
│       └── shell.go          # Shell hooks for Ctrl+R
├── go.mod
└── README.md
```

---

## 📜 License
MIT License. Built for personal data sovereignty and friction-free recall.
