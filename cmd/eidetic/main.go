package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ajejfiejof/eidetic/pkg/collectors"
	"github.com/ajejfiejof/eidetic/pkg/config"
	"github.com/ajejfiejof/eidetic/pkg/document"
	"github.com/ajejfiejof/eidetic/pkg/index"
	"github.com/ajejfiejof/eidetic/pkg/integration"
	"github.com/ajejfiejof/eidetic/pkg/tui"
)

const version = "1.0.0"

func printHelp() {
	fmt.Printf(`eidetic - Sovereign Cognitive Index (v%s)
Instant photographic recall across everything you've ever typed, copied, or read.

USAGE:
  eidetic                          Launch interactive TUI search
  eidetic search <query>           Search from the command line
  eidetic index                    Crawl and index shell history, clipboard, and notes
  eidetic add [text]               Add a snippet, note, or piped stdin
  eidetic watch                    Run background daemon to watch clipboard & history
  eidetic stats                    Display database and index statistics
  eidetic init <bash|zsh|fish>     Generate shell integration hook (Ctrl+R replacement)

OPTIONS:
  --query, -q <string>             Initial query for TUI or search
  --source <shell|clipboard|file>  Filter search results by source
  --limit <int>                    Maximum results to return (default: 20)
  --version, -v                    Show version information
  --help, -h                       Show this help message
`, version)
}

func main() {
	queryFlag := flag.String("query", "", "Search query")
	flag.StringVar(queryFlag, "q", "", "Search query (shorthand)")
	sourceFlag := flag.String("source", "", "Filter by source (shell, clipboard, file, note)")
	limitFlag := flag.Int("limit", 20, "Result limit")
	verFlag := flag.Bool("version", false, "Show version")
	flag.BoolVar(verFlag, "v", false, "Show version (shorthand)")

	flag.Usage = printHelp
	flag.Parse()

	if *verFlag {
		fmt.Printf("eidetic v%s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	storage, err := index.OpenStorage(cfg.DataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	engine, err := index.NewEngine(storage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing index engine: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "init":
		if len(args) < 2 {
			fmt.Println("Usage: eidetic init <bash|zsh|fish>")
			os.Exit(1)
		}
		hook, err := integration.GenerateShellHook(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(hook)

	case "stats":
		showStats(engine, cfg)

	case "add":
		handleAdd(args[1:], engine)

	case "index":
		handleIndex(engine, cfg)

	case "watch":
		handleWatch(engine, cfg)

	case "search":
		searchQuery := strings.Join(args[1:], " ")
		if searchQuery == "" {
			searchQuery = *queryFlag
		}
		handleSearch(engine, searchQuery, document.SourceType(*sourceFlag), *limitFlag)

	default:
		// Default: launch interactive TUI or run auto-index on first boot
		if engine.Count() == 0 {
			fmt.Println("⚡ First run detected: indexing shell history...")
			handleIndex(engine, cfg)
		}

		initialQuery := *queryFlag
		if cmd != "" {
			initialQuery = strings.Join(args, " ")
		}

		selected, err := tui.RunTUI(engine, initialQuery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
		if selected != nil {
			fmt.Print(selected.Content)
		}
	}
}

func showStats(engine *index.Engine, cfg *config.Config) {
	fmt.Println("================================================================================")
	fmt.Println("  EIDETIC: SOVEREIGN COGNITIVE INDEX STATISTICS")
	fmt.Println("================================================================================")
	fmt.Printf("Database Path : %s/documents.jsonl\n", cfg.DataDir)
	fmt.Printf("Total Items   : %d\n", engine.Count())
	fmt.Println("Breakdown by Source:")
	counts := engine.SourceCounts()
	for src, count := range counts {
		fmt.Printf("  • %-12s : %d\n", src, count)
	}
	fmt.Println("================================================================================")
}

func handleIndex(engine *index.Engine, cfg *config.Config) {
	start := time.Now()
	fmt.Println("🔍 Collecting shell history...")
	shellDocs, _ := collectors.CollectShellHistory()

	fmt.Println("📂 Crawling local notes and directories...")
	fileDocs, _ := collectors.CollectFiles(cfg.WatchDirs, cfg.MaxDocSizeKB, cfg.IgnoreRules)

	var clipDocs []document.Document
	if clipText, err := collectors.ReadClipboard(); err == nil && len(clipText) > 4 {
		clipDocs = append(clipDocs, document.NewDocument(document.SourceClipboard, clipText, "", time.Now()))
	}

	total := len(shellDocs) + len(fileDocs) + len(clipDocs)
	all := append(shellDocs, fileDocs...)
	all = append(all, clipDocs...)

	_ = engine.IndexBatch(all)
	elapsed := time.Since(start)

	fmt.Printf("✅ Indexed %d items in %v (Total in index: %d)\n",
		total, elapsed, engine.Count())
}

func handleAdd(args []string, engine *index.Engine) {
	var doc document.Document
	var err error

	if len(args) > 0 {
		content := strings.Join(args, " ")
		doc = collectors.CollectSnippet(content, "", nil)
	} else {
		// Read from piped stdin
		doc, err = collectors.CollectStdin("", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
	}

	if err := engine.Index(doc); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to index item: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Added item to Eidetic index (ID: %s)\n", doc.ID)
}

func handleSearch(engine *index.Engine, query string, source document.SourceType, limit int) {
	start := time.Now()
	results := engine.Search(query, index.SearchFilter{
		Source: source,
		Limit:  limit,
	})
	elapsed := time.Since(start)

	if len(results) == 0 {
		fmt.Printf("No results found for %q\n", query)
		return
	}

	fmt.Printf("Found %d matches in %.2fms:\n\n", len(results), float64(elapsed.Microseconds())/1000.0)
	for i, r := range results {
		fmt.Printf("[%2d] [%-6s] %s\n", i+1, r.Doc.Source, r.Doc.Title)
		lines := strings.Split(r.Doc.Content, "\n")
		for _, l := range lines {
			if len(l) > 90 {
				l = l[:87] + "..."
			}
			fmt.Printf("     %s\n", l)
		}
		fmt.Println()
	}
}

func handleWatch(engine *index.Engine, cfg *config.Config) {
	fmt.Println("🛡️  Eidetic daemon running: watching clipboard and system events...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go collectors.WatchClipboard(ctx, 1*time.Second, func(doc document.Document) {
		_ = engine.Index(doc)
		fmt.Printf("[+] Captured clipboard item (%d chars) at %s\n",
			len(doc.Content), doc.Timestamp.Format("15:04:05"))
	})

	<-sigChan
	fmt.Println("\nStopping daemon gracefully...")
}
