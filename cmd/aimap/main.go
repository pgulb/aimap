package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pgulb/aimap/internal/config"
	"github.com/pgulb/aimap/internal/output"
	"github.com/pgulb/aimap/internal/parser"
	"github.com/pgulb/aimap/internal/scanner"
	"github.com/pgulb/aimap/internal/symbol"
	"github.com/pgulb/aimap/internal/update"
)

// version is set at build time via -ldflags, e.g.:
// go build -ldflags="-X main.version=release-20260731-abc1234" ./cmd/aimap/
var version = "dev"

func main() {
	root := flag.String("path", ".", "project root directory")
	outputFlag := flag.String("output", "", "output file path (default MAP.md, or MAP_dev.md with --dev)")
	verbose := flag.Bool("v", false, "verbose output")
	dev := flag.Bool("dev", false, "use .aimapignore_dev and MAP_dev.md instead of production files")
	selfUpdate := flag.Bool("self-update", false, "check for updates and replace the current binary")
	enableSelfUpdate := flag.Bool("enable-self-update", false, "enable self-update for installations not using the install script")
	versionFlag := flag.Bool("version", false, "print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: aimap [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Scans Go and Python source files and generates MAP.md\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *versionFlag {
		fmt.Println("aimap", version)
		return
	}

	if *selfUpdate {
		if err := update.Do(version, *verbose); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *enableSelfUpdate {
		if update.CanSelfUpdate() {
			fmt.Println("Self-update is already enabled.")
			return
		}
		if err := update.CreateMarker(); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to enable self-update: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Self-update enabled. Run 'aimap --self-update' to update.")
		return
	}

	ignoreFileName := ".aimapignore"
	outPath := "MAP.md"
	if *dev {
		ignoreFileName = ".aimapignore_dev"
		outPath = "MAP_dev.md"
	}
	if *outputFlag != "" {
		outPath = *outputFlag
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "aimap: scanning %s\n", *root)
	}

	cfg, err := config.Load(*root, ignoreFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: loading config: %v\n", err)
		os.Exit(1)
	}

	if err := config.EnsureIgnoreFile(cfg.Root, ignoreFileName); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to ensure %s: %v\n", ignoreFileName, err)
	}

	s := scanner.NewScanner(cfg.Matcher)
	files, err := s.Scan(cfg.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: scanning files: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "aimap: found %d files to parse\n", len(files))
	}

	var allSymbols []symbol.Symbol
	for _, f := range files {
		if *verbose {
			fmt.Fprintf(os.Stderr, "aimap: parsing %s\n", f.Path)
		}
		syms, err := parser.ParseFile(f.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", f.Path, err)
			continue
		}
		allSymbols = append(allSymbols, syms...)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "aimap: writing %s\n", outPath)
	}

	if err := output.Render(allSymbols, cfg.Root, outPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: writing output: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "aimap: done\n")
	}
}
