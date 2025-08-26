// Package main provides the CLI entry point for contexture
package main

import (
	"os"

	"github.com/contextureai/contexture/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args))
}
