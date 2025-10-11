package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestNewNewCommand(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewNewCommand(deps)
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.projectManager)
	assert.NotNil(t, cmd.fs)
}

func TestNewAction_NoPath(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	app := createTestApp(func(ctx context.Context, cmd *cli.Command) error {
		return NewAction(ctx, cmd, deps)
	})

	err := runTestApp(app)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no path provided")
}

func TestNewCommand_CreateRuleOutsideProject(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewNewCommand(deps)

	// Create mock CLI command with flags
	cliCmd := &cli.Command{}
	ctx := context.Background()

	// Test creating rule outside project (no .contexture.yaml)
	// Note: The file will be created relative to the actual working directory
	// since Execute() uses os.Getwd(). We just test that no error occurs.
	err := cmd.Execute(ctx, cliCmd, "test-outside-rule")
	require.NoError(t, err)
}

func TestNewCommand_CreateRuleInsideProject(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewNewCommand(deps)

	// Create mock CLI command
	cliCmd := &cli.Command{}
	ctx := context.Background()

	// Test creating rule - just verify no error
	// E2E tests will verify the actual file creation behavior
	err := cmd.Execute(ctx, cliCmd, "my-project-rule")
	require.NoError(t, err)
}

func TestNewCommand_WithCustomMetadata(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewNewCommand(deps)

	// Create mock CLI command with custom flags
	cliCmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
			&cli.StringFlag{Name: "description"},
			&cli.StringFlag{Name: "tags"},
		},
	}

	// Set flag values
	flagSet := cliCmd.Flags
	for _, flag := range flagSet {
		if f, ok := flag.(*cli.StringFlag); ok {
			switch f.Name {
			case "name":
				f.Value = "Custom Rule Name"
			case "description":
				f.Value = "Custom rule description"
			case "tags":
				f.Value = "security,testing,custom"
			}
		}
	}

	ctx := context.Background()

	// Execute command - just verify no error
	err := cmd.Execute(ctx, cliCmd, "custom-rule-test")
	require.NoError(t, err)
}

func TestNewCommand_FileAlreadyExists(t *testing.T) {
	t.Parallel()
	// This test is better covered in E2E tests where we can control
	// the actual filesystem. Unit tests with mock fs can't easily test
	// this behavior since os.Getwd() returns the real directory.
	t.Skip("File existence checks are better tested in E2E tests")
}

func TestNewCommand_PathNormalization(t *testing.T) {
	t.Parallel()
	// Path normalization logic is tested in generateRuleContent and determineTargetPath
	// Full path behavior is tested in E2E tests
	t.Skip("Path normalization is better tested in E2E tests")
}

func TestParseTags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single tag",
			input:    "security",
			expected: []string{"security"},
		},
		{
			name:     "multiple tags",
			input:    "security,testing,auth",
			expected: []string{"security", "testing", "auth"},
		},
		{
			name:     "tags with spaces",
			input:    "security, testing, auth",
			expected: []string{"security", "testing", "auth"},
		},
		{
			name:     "tags with extra spaces",
			input:    "  security  ,  testing  ,  auth  ",
			expected: []string{"security", "testing", "auth"},
		},
		{
			name:     "empty tag in middle",
			input:    "security,,auth",
			expected: []string{"security", "auth"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCommand_GenerateRuleContent(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewNewCommand(deps)

	tests := []struct {
		name        string
		title       string
		description string
		tags        []string
		checkFor    []string
		notContain  []string
	}{
		{
			name:        "minimal rule - no optional fields",
			title:       "",
			description: "",
			tags:        nil,
			checkFor: []string{
				"---",
				"trigger: manual",
				"title: \"\"",
				"description: \"\"",
			},
			notContain: []string{
				"tags:",
				"# ",
			},
		},
		{
			name:        "basic rule",
			title:       "Test Rule",
			description: "This is a test rule",
			tags:        []string{"test"},
			checkFor: []string{
				"---",
				"title: Test Rule",
				"description: This is a test rule",
				"tags:",
				"- test",
				"trigger: manual",
				"# Test Rule",
			},
		},
		{
			name:        "rule with multiple tags",
			title:       "Security Rule",
			description: "Security check",
			tags:        []string{"security", "auth", "critical"},
			checkFor: []string{
				"title: Security Rule",
				"- security",
				"- auth",
				"- critical",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := cmd.generateRuleContent(tt.title, tt.description, tt.tags)
			require.NoError(t, err)
			assert.NotEmpty(t, content)

			for _, check := range tt.checkFor {
				assert.Contains(t, content, check, "Content should contain: %s", check)
			}

			for _, notCheck := range tt.notContain {
				assert.NotContains(t, content, notCheck, "Content should not contain: %s", notCheck)
			}

			// Verify YAML frontmatter structure
			assert.True(t, strings.HasPrefix(content, "---\n"))
			parts := strings.Split(content, "\n---\n")
			assert.Len(t, parts, 2, "Content should have frontmatter and body")
		})
	}
}

func TestNewCommand_DetermineTargetPath_ProjectContext(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		configLocation   domain.ConfigLocation
		configPath       string
		rulePath         string
		expectedContains string
	}{
		{
			name:             "root config location",
			configLocation:   domain.ConfigLocationRoot,
			configPath:       "/project/.contexture.yaml",
			rulePath:         "my-rule",
			expectedContains: "/rules/my-rule.md",
		},
		{
			name:             "contexture dir config location",
			configLocation:   domain.ConfigLocationContexture,
			configPath:       "/project/.contexture/.contexture.yaml",
			rulePath:         "my-rule",
			expectedContains: "/rules/my-rule.md",
		},
		{
			name:             "nested rule path",
			configLocation:   domain.ConfigLocationRoot,
			configPath:       "/project/.contexture.yaml",
			rulePath:         "security/auth",
			expectedContains: "/rules/security/auth.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			// Create config file
			configDir := "/project"
			if tt.configLocation == domain.ConfigLocationContexture {
				configDir = "/project/.contexture"
			}
			_ = fs.MkdirAll(configDir, 0o755)

			configContent := `version: 1
formats:
  - type: claude
    enabled: true
rules: []
`
			_ = afero.WriteFile(fs, tt.configPath, []byte(configContent), 0o644)

			deps := &dependencies.Dependencies{
				FS:      fs,
				Context: context.Background(),
			}

			cmd := NewNewCommand(deps)

			// Determine target path
			targetPath := cmd.determineTargetPath("/project", tt.rulePath)
			assert.Contains(t, targetPath, tt.expectedContains)
		})
	}
}

func TestNewCommand_DetermineTargetPath_NoProject(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	tempDir := "/tmp/no-project"
	_ = fs.MkdirAll(tempDir, 0o755)

	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewNewCommand(deps)

	tests := []struct {
		name         string
		rulePath     string
		expectedPath string
	}{
		{
			name:         "simple path",
			rulePath:     "my-rule",
			expectedPath: tempDir + "/my-rule.md",
		},
		{
			name:         "nested path",
			rulePath:     "security/auth",
			expectedPath: tempDir + "/security/auth.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetPath := cmd.determineTargetPath(tempDir, tt.rulePath)
			assert.Equal(t, tt.expectedPath, targetPath)
		})
	}
}

func TestNewCommand_CreateNestedDirectories(t *testing.T) {
	t.Parallel()
	// Nested directory creation is better tested in E2E tests
	// where we can control the actual filesystem
	t.Skip("Nested directory creation is better tested in E2E tests")
}
