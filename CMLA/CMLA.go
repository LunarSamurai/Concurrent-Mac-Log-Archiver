package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	printTitleBanner()
	if len(os.Args) < 2 || os.Args[1] != "collect" {
		fmt.Println("Usage: log collect --input-dir <source> --macos <version> [--output <dest>.logarchive]")
		return
	}

	cmd := flag.NewFlagSet("collect", flag.ExitOnError)
	input := cmd.String("input-dir", "", "Root of the source file tree (private/var, Users)")
	output := cmd.String("output", "recovered.logarchive", "Destination .logarchive path")
	macos := cmd.String("macos", "", "Source macOS version (e.g. 10.13, 10.14, 10.15, 12.0)")
	cmd.Parse(os.Args[2:])

	if *input == "" || *macos == "" {
		fmt.Println("Error: --input-dir and --macos are required")
		os.Exit(1)
	}

	ver, err := mapOSArchiveVersion(*macos)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := collectFromInputDir(*input, *output, *macos, ver); err != nil {
		fmt.Println("Collection failed:", err)
		os.Exit(1)
	}

	fmt.Printf("Log archive created at: %s (macOS %s → OSArchiveVersion=%d)\n", *output, *macos, ver)
	printProcessCompleteBanner()
	printTitleBanner()
}

func mapOSArchiveVersion(macos string) (int, error) {
	// Based on mapping observed in live log collect Info.plist files :contentReference[oaicite:1]{index=1}
	// macOS Sierra (10.12) → archiveVersion 3
	// macOS Mojave / Catalina (10.14 / 10.15) → archiveVersion 4
	// Monterey onward (12.x, 13.x, 14.x, 15.x/Sequoia up to 15.6) → 5 :contentReference[oaicite:2]{index=2}

	// Normalize version string
	v := strings.TrimSpace(macos)

	// 10.12–10.13 → 3
	if strings.HasPrefix(v, "10.12") || strings.HasPrefix(v, "10.13") {
		return 3, nil
	}
	// 10.14–10.15 → 4
	if strings.HasPrefix(v, "10.14") || strings.HasPrefix(v, "10.15") {
		return 4, nil
	}
	// 12.x, 13.x, 14.x, 15.x, 16.x → 5
	majorMinor := strings.Split(v, ".")
	if len(majorMinor) >= 1 {
		base := majorMinor[0]
		switch base {
		case "12", "13", "14", "15", "16":
			return 5, nil
		}
	}
	return 0, fmt.Errorf("unsupported or unknown macOS version: %s", macos)
}

func collectFromInputDir(input, outputPath, osVersion string, archiveVer int) error {
	srcMap := map[string]string{
		"diagnostics":      filepath.Join(input, "var", "db", "diagnostics", "**"),
		"uuidtext":         filepath.Join(input, "var", "db", "uuidtext", "**"),
		"timesync":         filepath.Join(input, "var", "db", "diagnostics", "timesync"),
		"system_logs":      filepath.Join(input, "var", "log"),
		"user_logs":        filepath.Join(input, "Library", "Logs"),
		"live":             filepath.Join(input, "var", "db", "logd", "streams"),
		"LogStoreMetadata": filepath.Join(input, "var", "db", "logd"),
		"network":          filepath.Join(input, "var", "log", "DiagnosticMessages"),
	}

	if !strings.HasSuffix(outputPath, ".logarchive") {
		outputPath += ".logarchive"
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}

	fmt.Println("Creating .logarchive at", outputPath)
	for label, src := range srcMap {
		dst := filepath.Join(outputPath, label)

		if strings.Contains(src, "**") {
			matches, err := filepath.Glob(src)
			if err != nil {
				fmt.Printf("Warning: invalid pattern %s: %v\n", src, err)
				continue
			}
			for _, srcRecursive := range matches {
				base := filepath.Base(srcRecursive)
				dstMatch := filepath.Join(outputPath, base)
				fmt.Printf("Copying (wildcard) %s -> %s\n", srcRecursive, dstMatch)
				if err := copyDirectory(srcRecursive, dstMatch); err != nil {
					fmt.Printf("Warning: failed to copy %s: %v\n", srcRecursive, err)
				}
			}
		} else {
			fmt.Printf("Copying %s -> %s\n", src, dst)
			if err := copyDirectory(src, dst); err != nil {
				fmt.Printf("Warning: failed to copy %s: %v\n", label, err)
			}
		}
	}

	plist := buildPlist(osVersion, archiveVer)
	if err := os.WriteFile(filepath.Join(outputPath, "Info.plist"), []byte(plist), 0644); err != nil {
		return err
	}

	return nil
}

func copyDirectory(src, dest string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

func buildPlist(osVersion string, archiveVer int) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>OSVersion</key><string>macOS %s</string>
  <key>LogArchiveVersion</key><integer>1</integer>
  <key>OSArchiveVersion</key><integer>%d</integer>
  <key>Collected</key><true/>
  <key>TimeCreated</key>
  <date>%s</date>
</dict>
</plist>`, osVersion, archiveVer, time.Now().UTC().Format("2006-01-02T15:04:05Z"))
}

// ============================ ASCII Art Section ============================

func printTitleBanner() {
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                   ███╗   ███╗ █████╗  ██████╗ ██████╗ ███████╗                       │")
	fmt.Println("│                   ████╗ ████║██╔══██╗██╔════╝██╔═══██╗██╔════╝                       │")
	fmt.Println("│                   ██╔████╔██║███████║██║     ██║   ██║███████╗                       │")
	fmt.Println("│                   ██║╚██╔╝██║██╔══██║██║     ██║   ██║╚════██║                       │")
	fmt.Println("│                   ██║ ╚═╝ ██║██║  ██║╚██████╗╚██████╔╝███████║                       │")
	fmt.Println("│                   ╚═╝     ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝                       │")
	fmt.Println("│      __                     ___                    __                                │")
	fmt.Println("│     / /   ____    ____ _   /   |   _____  _____   / /_    (_) _   __  ___    _____   │")
	fmt.Println("│    / /   / __ \\  / __ `/  / /| |  / ___/ / ___/  / ___\\  / / | | / / / _ \\  / ___/   │")
	fmt.Println("│   / /___/ /_/ / / /_/ /  / ___ | / /    / /__   / / / / / /  | |/ / /  __/ / /       │")
	fmt.Println("│  /_____/ \\___/  \\_,  /  /_/  |_|/_/     \\___/  /_/ /_/ /_/   |___/  \\___/ /_/        │")
	fmt.Println("│               /____/                                                                 │")
	fmt.Println("│                                                                                      │")
	fmt.Println("│                                Deadbox Edition                                       │")
	fmt.Println("│                                 Version 1.0.0                                        │")
	fmt.Println("│                            github.com/LunarSamurai                                   │")
	fmt.Println("└──────────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println(`
                                              |
                                             |  |
                                              |
                                         _ /_
                                    |   ( '' )
                                     |   '~~'
                                    |         |
                                      _ /_   |  |                
                                     ( '' )  _\ _
                                 __---'~~'--( '' )--__
                                |||||||||||||||||||||||
                                 |  _ _ _  __   ___  |
                                 |  \_|_/  __|_   |  |
                                  \_________________/
                                                `)
}

func printProcessCompleteBanner() {
	const (
		green = "\033[32m" // ANSI code for green text
		reset = "\033[0m"  // Resets colors to the terminal default
	)

	processCompleteBanner := `
====================================================================================================
        ____                                    ______                      __     __     
       / __ \_________  ________  __________   / ____/___  ____ ___  ____  / /__  / /____ 
      / /_/ / ___/ __ \/ ___/ _ \/ ___/ ___/  / /   / __ \/ __ '__ \/ __ \/ / _ \/ __/ _ \
     / ____/ /  / /_/ / /__/  __(__  |__  )  / /___/ /_/ / / / / / / /_/ / /  __/ /_/  __/
    /_/   /_/   \____/\___/\___/____/____/   \____/\____/_/ /_/ /_/ .___/_/\___/\__/\___/ 
                                                                 /_/                      
====================================================================================================																 
	`

	fmt.Println("\n\n" + green + processCompleteBanner + reset + "\n\n")
}
