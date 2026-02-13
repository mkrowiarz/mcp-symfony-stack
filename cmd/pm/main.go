package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/mcp"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/tui"
)

func main() {
	mcpFlag := flag.Bool("mcp", false, "Run as MCP server (stdio transport)")
	flag.Parse()

	if *mcpFlag {
		if err := mcp.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	args := flag.Args()
	if len(args) > 0 && args[0] == "init" {
		writeFlag := false
		for _, arg := range args[1:] {
			if arg == "--write" || arg == "-w" {
				writeFlag = true
			}
		}

		result, err := commands.Init(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if writeFlag {
			configDir := ".haive"
			configPath := filepath.Join(configDir, "config.json")

			if _, err := os.Stat(configPath); err == nil {
				fmt.Fprintf(os.Stderr, "Config file already exists: %s\n", configPath)
				fmt.Fprintf(os.Stderr, "Remove it first or use 'pm init' to preview.\n")
				os.Exit(1)
			}

			if err := os.MkdirAll(configDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
				os.Exit(1)
			}

			if err := os.WriteFile(configPath, []byte(result.SuggestedConfig), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Created: %s\n", configPath)
		} else {
			fmt.Println(result.SuggestedConfig)
		}
		return
	}

	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
