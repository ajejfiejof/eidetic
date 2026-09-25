# Eidetic 🧠 🏳️‍⚧️

> **Your sovereign cognitive index — instant photographic recall across everything you've ever typed, copied, committed, or read.**

`eidetic` is a single-binary, local-first search engine and cognitive memory layer for Linux developers. It automatically captures, sanitizes, indexes, and ranks your shell history, clipboard history, git commits, notes, and code snippets into an ultra-fast local index, accessible via an interactive terminal UI or instant CLI.

Zero cloud. Zero telemetry. Zero external database services. Crafted with clean Go, mechanical sympathy, and pride.

---

## ⚡ Key Features

* **Sub-Millisecond Search**: In-memory inverted index powered by **BM25 relevance scoring**, exact phrase matching, and **Levenshtein fuzzy matching** for automatic typo recovery.
* **Interactive Split-Screen TUI**: As-you-type fuzzy searching built with Bubbletea & Lipgloss. Live preview pane, relative timestamps, and one-key selection/yank.
* **Multi-Shell History Ingestion**: Automatically discovers and indexes `.bash_history`, `.zsh_history`, and Fish shell history with timestamps and secret sanitization.
* **Git Repository & Commit Indexer**: Crawls your local project repositories (e.g. `~/Projects`) to index commits, hashes, authors, and commit messages (`--source git`).
* **Clipboard Watcher & Systemd User Daemon**: Background clipboard monitoring on both **Wayland** (`wl-paste`) and **X11** (`xclip`/`xsel`).
* **Pinning & Bookmarks (`⭐`)**: Pin your favorite or most critical commands and snippets so they always rise to the top of search results.
* **Theme Engine**: Beautiful visual themes including **Trans Pride** (default pastel pink, cyan & white palette), **Catppuccin Mocha**, and **Cyber Default**.
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
# Indexes shell history, git commits, clipboard, and local notes
eidetic index
```

### 3. Search
```bash
# Launch interactive TUI
eidetic

# Or search from the command line
eidetic search docker
eidetic search --source git "commit message"
```

---

## ⌨️ Interactive TUI Navigation

Running `eidetic` without arguments opens the full-screen terminal interface:

| Keybinding | Action |
|---|---|
| `Type any text` | Instant as-you-type fuzzy search |
| `↑` / `Ctrl+K` / `Ctrl+P` | Move selection up |
| `↓` / `Ctrl+J` / `Ctrl+N` | Move selection down |
| `Tab` | Cycle source filter (`ALL` → `SHELL` → `CLIP` → `GIT` → `FILE` → `NOTE`) |
| `Ctrl+B` | **Pin / Unpin** currently highlighted item (`⭐`) |
| `Enter` | Select item (automatically copies to clipboard & inserts into shell) |
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
eidetic search --source git "kangaroo"
eidetic search --source clip "curl"
```

### Pinning Items
```bash
# Pin an item so it stays at the top of results
eidetic pin <doc_id>
```

### Background Daemon (Systemd User Service)
```bash
# Install and enable background clipboard daemon
eidetic service install

# Check status
eidetic service status

# Uninstall
eidetic service uninstall
```

### Manual Note & Stdin Capture
```bash
# Add a quick snippet
eidetic add "openssl s_client -connect example.com:443 -servername example.com"

# Pipe output from any command
find . -type f -name '*.go' | xargs wc -l | eidetic add
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

## 🎨 Theme Customization

Eidetic includes built-in themes configured in `~/.config/eidetic/config.json`:

```json
{
  "theme": "trans",
  "auto_copy": true,
  "watch_dirs": [
    "/home/ashley/Notes",
    "/home/ashley/.config"
  ],
  "git_dirs": [
    "/home/ashley/Projects"
  ]
}
```

* `"trans"`: Pastel Pink (`#F5A9B8`), Pastel Cyan (`#5BCEFA`), and White (`#FFFFFF`).
* `"catppuccin"`: Catppuccin Mocha palette (`#CBA6F7`, `#A6E3A1`, `#89DCEB`).
* `"default"`: Cyberpunk purple and green.

---

## 📊 Performance (Tested on ThinkPad T440p)

* **Startup & Query Latency**: **10 to 100 microseconds** ($0.01\text{ms} - 0.10\text{ms}$) across 3,500+ items.
* **Index Throughput**: **4,090+ items indexed in seconds**.
* **Memory Footprint**: $<18\text{MB}$ resident set size.
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
│   │   └── document.go       # Document model, deterministic ID & pinning
│   ├── clipboard/
│   │   └── clipboard.go      # Cross-desktop clipboard copying
│   ├── index/
│   │   ├── engine.go         # Inverted index, BM25 scoring & fuzzy matching
│   │   ├── tokenizer.go      # Code & shell aware tokenizer
│   │   ├── storage.go        # Append-only JSONL persistence
│   │   └── engine_test.go    # Unit tests for search & ranking
│   ├── collectors/
│   │   ├── shell.go          # Bash, Zsh, Fish history parsers
│   │   ├── git.go            # Git repository & commit crawler
│   │   ├── clipboard.go      # Wayland & X11 clipboard watcher
│   │   ├── files.go          # Notes & markdown crawler
│   │   └── stdin.go          # Piped input ingestion
│   ├── tui/
│   │   ├── model.go          # Bubbletea TUI model & state loop
│   │   └── styles.go         # Lipgloss themes (Trans Pride, Catppuccin)
│   └── integration/
│       ├── shell.go          # Shell hooks for Ctrl+R
│       └── systemd.go        # User systemd service installer
├── go.mod
└── README.md
```

---

## 📜 License
MIT License. Built for personal data sovereignty, instant recall, and developer joy.
