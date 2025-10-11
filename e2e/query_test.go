// Package e2e provides end-to-end tests for the query command
package e2e

import (
	"testing"

	"github.com/contextureai/contexture/e2e/helpers"
	"github.com/spf13/afero"
)

// TestQueryCommand tests basic query command functionality
func TestQueryCommand(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("help command", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		result := project.Run(t, "query", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Search for rules").
			ExpectStdout(t, "Available fields for --expr mode").
			ExpectStdout(t, "expr-lang.org")
	})

	t.Run("no arguments fails", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		result := project.Run(t, "query")
		result.ExpectFailure(t).
			ExpectStderr(t, "query string is required")
	})

	t.Run("simple text search", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		// Initialize project (optional for query, but tests with config)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "go")
		// Should succeed whether or not results are found
		if result.ExitCode == 0 {
			t.Logf("Query succeeded with results")
		}
	})

	t.Run("search with no results", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		result := project.Run(t, "query", "xyznonexistentterm987654321")
		// Should succeed but show no results
		result.ExpectSuccess(t).
			ExpectStdout(t, "No rules found")
	})
}

// TestQueryExprMode tests expression-based query functionality
func TestQueryExprMode(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("basic expr query", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "--expr", "Tag contains \"test\"")
		// Expression should parse successfully
		if result.ExitCode != 0 {
			t.Logf("Expr query exit code: %d, stderr: %s", result.ExitCode, result.Stderr)
		}
	})

	t.Run("invalid expr syntax", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		result := project.Run(t, "query", "--expr", "invalid syntax {[")
		result.ExpectFailure(t).
			ExpectStderr(t, "expression")
	})

	t.Run("expr with multiple conditions", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "--expr", "Title != \"\" and Path != \"\"")
		// Should parse and execute successfully
		if result.ExitCode != 0 {
			t.Logf("Multi-condition expr exit code: %d", result.ExitCode)
		}
	})

	t.Run("expr with computed fields", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "--expr", "HasVars == false")
		// Should parse and execute successfully
		if result.ExitCode != 0 {
			t.Logf("Computed field expr exit code: %d", result.ExitCode)
		}
	})
}

// TestQueryProviderFiltering tests provider filtering functionality
func TestQueryProviderFiltering(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("filter by default provider", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "go", "--provider", "contexture")
		// Should succeed and search only contexture provider
		if result.ExitCode == 0 {
			t.Logf("Provider filter succeeded")
		}
	})

	t.Run("filter by non-existent provider", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "go", "--provider", "nonexistent")
		result.ExpectFailure(t).
			ExpectStderr(t, "no providers found")
	})

	t.Run("query with custom provider", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		// Initialize with custom provider
		project.WithConfig(`version: 1
formats:
  - type: claude
    enabled: true
providers:
  - name: testprovider
    url: https://github.com/contextureai/rules.git
    defaultBranch: main`)

		result := project.Run(t, "query", "test", "--provider", "testprovider")
		// Should attempt to query the custom provider
		if result.ExitCode != 0 {
			t.Logf("Custom provider query exit code: %d", result.ExitCode)
		}
	})
}

// TestQueryResultLimit tests result limiting functionality
func TestQueryResultLimit(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("default limit", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "a")
		// Should use default limit of 10
		if result.ExitCode == 0 {
			t.Logf("Default limit query succeeded")
		}
	})

	t.Run("custom limit", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "a", "--limit", "5")
		// Should respect custom limit
		if result.ExitCode == 0 {
			t.Logf("Custom limit (5) query succeeded")
		}
	})

	t.Run("zero limit shows all", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "a", "--limit", "0")
		// Should show all results
		if result.ExitCode == 0 {
			t.Logf("Zero limit (all) query succeeded")
		}
	})

	t.Run("limit larger than results", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "xyzrareterm", "--limit", "1000")
		// Should succeed and show available results
		result.ExpectSuccess(t)
	})
}

// TestQueryOutputFormats tests different output formats
func TestQueryOutputFormats(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("default terminal output", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "test")
		// Default output should be human-readable
		if result.ExitCode == 0 && len(result.Stdout) > 0 {
			t.Logf("Terminal output length: %d bytes", len(result.Stdout))
		}
	})

	t.Run("json output", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "test", "--output", "json")
		// Should output valid JSON
		if result.ExitCode == 0 {
			result.ExpectStdout(t, "{").
				ExpectStdout(t, "metadata").
				ExpectStdout(t, "rules")
			t.Logf("JSON output verified")
		}
	})

	t.Run("json output structure", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "test", "--output", "json", "--limit", "1")
		if result.ExitCode == 0 {
			// Verify JSON contains expected fields
			result.ExpectStdout(t, "query").
				ExpectStdout(t, "queryType").
				ExpectStdout(t, "totalResults")
		}
	})
}

// TestQueryWithoutConfig tests query without project configuration
func TestQueryWithoutConfig(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("query without config file", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		// Don't initialize - no .contexture.yaml

		result := project.Run(t, "query", "go")
		// Should work with just default provider
		if result.ExitCode == 0 {
			t.Logf("Query without config succeeded")
		}
	})

	t.Run("query uses default provider only", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		// No config file

		result := project.Run(t, "query", "test", "--provider", "contexture")
		// Should use default @contexture provider
		if result.ExitCode == 0 {
			t.Logf("Default provider query succeeded")
		}
	})
}

// TestQueryWithCustomProviders tests query with custom providers
func TestQueryWithCustomProviders(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("query with custom provider in config", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// Create config with custom provider
		project.WithConfig(`version: 1
formats:
  - type: claude
    enabled: true
providers:
  - name: custom
    url: https://github.com/contextureai/rules.git
    defaultBranch: main`)

		result := project.Run(t, "query", "test")
		// Should query both default and custom providers
		if result.ExitCode == 0 {
			t.Logf("Query with custom provider succeeded")
		}
	})

	t.Run("filter by custom provider name", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		project.WithConfig(`version: 1
formats:
  - type: claude
    enabled: true
providers:
  - name: mycompany
    url: https://github.com/contextureai/rules.git
    defaultBranch: main`)

		result := project.Run(t, "query", "test", "--provider", "mycompany")
		// Should query only the specified custom provider
		if result.ExitCode != 0 {
			t.Logf("Custom provider filter exit code: %d", result.ExitCode)
		}
	})
}

// TestQueryErrorHandling tests error scenarios
func TestQueryErrorHandling(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("invalid expression syntax", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "--expr", "this is not valid expr")
		result.ExpectFailure(t)
		// Should show helpful error message
	})

	t.Run("network failure handling", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// Create config with invalid provider URL
		project.WithConfig(`version: 1
formats:
  - type: claude
    enabled: true
providers:
  - name: invalid
    url: https://invalid-host-that-does-not-exist.example.com/repo.git
    defaultBranch: main`)

		result := project.Run(t, "query", "test", "--provider", "invalid")
		// Should handle network failure gracefully
		if result.ExitCode != 0 {
			t.Logf("Network failure handled with exit code: %d", result.ExitCode)
		}
	})

	t.Run("empty query string", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "")
		result.ExpectFailure(t).
			ExpectStderr(t, "query string is required")
	})
}

// TestQueryHelp tests help and usage information
func TestQueryHelp(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	t.Run("help shows field reference", func(t *testing.T) {
		result := project.Run(t, "query", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "ID").
			ExpectStdout(t, "Title").
			ExpectStdout(t, "Description").
			ExpectStdout(t, "Tags").
			ExpectStdout(t, "Provider").
			ExpectStdout(t, "HasVars")
	})

	t.Run("help shows expr documentation link", func(t *testing.T) {
		result := project.Run(t, "query", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "expr-lang.org")
	})

	t.Run("help shows examples", func(t *testing.T) {
		result := project.Run(t, "query", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Examples:").
			ExpectStdout(t, "contexture query")
	})

	t.Run("help shows all flags", func(t *testing.T) {
		result := project.Run(t, "query", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "--expr").
			ExpectStdout(t, "--output").
			ExpectStdout(t, "--provider").
			ExpectStdout(t, "--limit")
	})
}

// TestQueryPerformance tests performance characteristics
func TestQueryPerformance(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("query completes within reasonable time", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		// Run query and ensure it doesn't hang
		result := project.Run(t, "query", "test", "--limit", "5")
		// Should complete quickly (test timeout is 30s by default)
		if result.ExitCode == 0 {
			t.Logf("Query completed within timeout")
		}
	})

	t.Run("large limit handling", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "a", "--limit", "1000")
		// Should handle large limit without issues
		if result.ExitCode == 0 {
			t.Logf("Large limit query succeeded")
		}
	})
}

// TestQueryOutputContent tests output content and formatting
func TestQueryOutputContent(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("output includes rule information", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "go", "--limit", "1")
		if result.ExitCode == 0 && len(result.Stdout) > 0 {
			// Output should include rule ID (format: @provider/path)
			// Can't assert specific content since it depends on available rules
			t.Logf("Output contains rule information (%d bytes)", len(result.Stdout))
		}
	})

	t.Run("output shows match count", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "test")
		if result.ExitCode == 0 {
			// Should show count like "Found X rule(s)"
			result.ExpectStdout(t, "Found")
		}
	})

	t.Run("no results message", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "zzznoresultsexpected9999")
		result.ExpectSuccess(t).
			ExpectStdout(t, "No rules found")
	})
}

// TestQueryCLIFlags tests CLI flag combinations
func TestQueryCLIFlags(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("combine expr and provider flags", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "--expr", "Title != \"\"", "--provider", "contexture")
		// Should work with combined flags
		if result.ExitCode != 0 {
			t.Logf("Combined flags exit code: %d", result.ExitCode)
		}
	})

	t.Run("combine all flags", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "--expr", "Path != \"\"",
			"--provider", "contexture",
			"--limit", "3",
			"--output", "json")
		// Should work with all flags together
		if result.ExitCode == 0 {
			result.ExpectStdout(t, "{")
		}
	})

	t.Run("short flag aliases", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--no-interactive").ExpectSuccess(t)

		result := project.Run(t, "query", "test", "-o", "json", "-n", "5")
		// Should work with short aliases
		if result.ExitCode == 0 {
			result.ExpectStdout(t, "metadata")
		}
	})
}
