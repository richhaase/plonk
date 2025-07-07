//go:build ignore

// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package main

import (
	"fmt"
	"os"

	"plonk/internal/tasks"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	var err error
	switch command {
	case "build":
		err = tasks.Build()
	case "test":
		err = tasks.Test()
	case "precommit":
		err = tasks.Precommit()
	case "clean":
		err = tasks.Clean()
	case "install":
		err = tasks.Install()
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("dev.go - Pure Go development task runner")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run dev.go <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  build     Build the plonk binary")
	fmt.Println("  test      Run all tests")
	fmt.Println("  precommit Run pre-commit checks (format, lint, test, security)")
	fmt.Println("  clean     Clean build artifacts")
	fmt.Println("  install   Install plonk globally (go install)")
	fmt.Println("  help      Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run dev.go build")
	fmt.Println("  go run dev.go precommit")
	fmt.Println("  go run dev.go install")
}
