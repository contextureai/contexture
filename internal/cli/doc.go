// Package cli provides command-line interface utilities for the Contexture application.
// It includes help formatting, command rendering, and user interaction components.
//
// The package offers customized help printing with color support and improved formatting
// for urfave/cli commands. It integrates with the internal UI package for consistent
// theming across the application.
//
// Example:
//
//	// Set a custom help printer for your CLI app
//	app := &cli.Command{
//	    Name: "myapp",
//	    HelpPrinter: cli.NewHelpPrinter(),
//	}
//
//	// Or use the colored help printer directly
//	printer := cli.NewHelpPrinter()
//	err := printer.Print(os.Stdout, template, data)
package cli
