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
	printBanner()
	if len(os.Args) < 2 {
		fmt.Println("Usage: log collect [--output output_dir.logarchive] [--input-dir files_req]")
		return
	}

	if os.Args[1] == "collect" {
		collectCmd := flag.NewFlagSet("collect", flag.ExitOnError)
		output := collectCmd.String("output", "output.logarchive", "Directory to save the .logarchive")
		inputDir := collectCmd.String("input-dir", "", "Path to folder with full /private and /Users tree")
		collectCmd.Parse(os.Args[2:])

		if *inputDir == "" {
			fmt.Println("Error: --input-dir is required for deadbox collection.")
			os.Exit(1)
		}

		err := collectFromInputDir(*inputDir, *output)
		if err != nil {
			fmt.Println("Collection failed:", err)
		}
	} else {
		fmt.Println("Unknown command:", os.Args[1])
	}
}

func collectFromInputDir(inputDir string, outputDir string) error {
	fmt.Println("Using --input-dir mode to collect logs from deadbox folder")

	srcMap := map[string]string{
		"uuidtext":         filepath.Join(inputDir, "private", "var", "db", "uuidtext"),
		"system_logs":      filepath.Join(inputDir, "private", "var", "log"),
		"diagnostics":      filepath.Join(inputDir, "private", "var", "db", "diagnostics"),
		"user_logs":        filepath.Join(inputDir, "Users"),
		"timesync":         filepath.Join(inputDir, "private", "var", "db", "timesync"),
		"live":             filepath.Join(inputDir, "private", "var", "db", "logd", "streams"),
		"LogStoreMetadata": filepath.Join(inputDir, "private", "var", "db", "logd"),
		"network":          filepath.Join(inputDir, "private", "var", "log", "DiagnosticMessages"),
	}

	logarchivePath := filepath.Clean(outputDir)
	if !strings.HasSuffix(logarchivePath, ".logarchive") {
		logarchivePath += ".logarchive"
	}
	os.MkdirAll(logarchivePath, 0755)

	for name, src := range srcMap {
		dst := filepath.Join(logarchivePath, name)
		fmt.Printf("Copying %s → %s\n", src, dst)
		err := copyDirectory(src, dst)
		if err != nil {
			fmt.Printf("Failed to copy %s: %v\n", name, err)
		}
	}

	infoPlistPath := filepath.Join(logarchivePath, "Info.plist")
	return os.WriteFile(infoPlistPath, []byte(examplePlist()), 0644)
}

func copyDirectory(src, dest string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(path, src)
		targetPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}

func examplePlist() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>OSVersion</key>
	<string>macOS 12.0</string>
	<key>LogArchiveVersion</key>
	<integer>1</integer>
	<key>Collected</key>
	<true/>
	<key>TimeCreated</key>
	<date>` + time.Now().UTC().Format("2006-01-02T15:04:05Z") + `</date>
</dict>
</plist>`
}

func printBanner() string {
	return `
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
	`
}

func printTitleBanner() {
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│	         	     ███╗   ███╗ █████╗  ██████╗ ██████╗ ███████╗         			 	│")
	fmt.Println("│	         	     ████╗ ████║██╔══██╗██╔════╝██╔═══██╗██╔════╝         			 	│")
	fmt.Println("│	           	     ██╔████╔██║███████║██║     ██║   ██║███████╗         			 	│")
	fmt.Println("│	         	     ██║╚██╔╝██║██╔══██║██║     ██║   ██║╚════██║						│")
	fmt.Println("│  		 	     ██║ ╚═╝ ██║██║  ██║╚██████╗╚██████╔╝███████║						│")
	fmt.Println("│				     ╚═╝     ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝					 	│")
	fmt.Println("│	    __                     ___                    __      _                      	│")
	fmt.Println("│	   / /   ____    ____ _   /   |   _____  _____   / /_    (_) _   __  ___    _____	│")
	fmt.Println("│	  / /   / __\\  / __ `/  / /| |  / ___/ / ___/  / __ \\ / / | | / / / _ \\ / ___/	│")
	fmt.Println("│	 / /___/ /_/ / / /_/ /  / ___ | / /    / /__   / / / / / /  | |/ / /  __/ / /    	│")
	fmt.Println("│	/_____/\\___/  \\_, /  /_/  |_|/_/     \\__/  /_/ /_/ /_/   |___/  \\__/ /_/     	│")
	fmt.Println("│		          /____/                                                             	│")
	fmt.Println("│    			   		                                            				 	│")
	fmt.Println("│    			   		          Deadbox Edition                     				 	│")
	fmt.Println("│          		 		  github.com/LunarSamurai   		        			 	│")
	fmt.Println("└──────────────────────────────────────────────────────────────────────────────────────┘")
}
