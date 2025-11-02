package rule

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/spf13/afero"
	"github.com/titanous/json5"
)

// Compile regex patterns once
var localVariablesPattern = domain.VariablesPatternRegex

// LocalFetcher implements rule fetching from local filesystem
type LocalFetcher struct {
	fs      afero.Fs
	baseDir string
	parser  Parser
}

// NewLocalFetcher creates a fetcher that reads rules from local filesystem
func NewLocalFetcher(fs afero.Fs, baseDir string) *LocalFetcher {
	return &LocalFetcher{
		fs:      fs,
		baseDir: baseDir,
		parser:  NewParser(),
	}
}

// ParseRuleID parses a local rule ID (simplified format or full format)
func (f *LocalFetcher) ParseRuleID(ruleID string) (*domain.ParsedRuleID, error) {
	// Handle full format [contexture(local):path] or [contexture(local):path,ref]{variables}
	if strings.HasPrefix(ruleID, "[contexture(local):") {
		// Use the domain-level rule ID parser for full format
		matches := domain.RuleIDParsePatternRegex.FindStringSubmatch(ruleID)
		if len(matches) > 0 {
			path := matches[2]
			parsed := &domain.ParsedRuleID{
				RulePath: path, // Path component
				Source:   "local",
				Ref:      "", // Local fetcher doesn't use refs
			}

			// Parse optional variables from the full format
			if len(matches) > 4 && matches[4] != "" {
				variables := make(map[string]any)
				if err := json5.Unmarshal([]byte(matches[4]), &variables); err != nil {
					return nil, contextureerrors.WithOpf("ParseRuleID", "invalid JSON5 variables in rule ID '%s': %w", ruleID, err)
				}
				parsed.Variables = variables
			}

			return parsed, nil
		}
	}

	// Handle simplified format: "path/to/rule{variables}" or just "path/to/rule"
	matches := localVariablesPattern.FindStringSubmatch(ruleID)

	if len(matches) == 0 {
		return nil, contextureerrors.WithOpf("ParseRuleID", "invalid local rule ID format: %s", ruleID)
	}

	parsed := &domain.ParsedRuleID{
		RulePath: matches[1], // Path component
		Source:   "local",
		Ref:      "", // Local fetcher doesn't use refs
	}

	// Parse optional variables
	if len(matches) > 2 && matches[2] != "" {
		variablesJSON := matches[2]
		variables := make(map[string]any)
		if err := json5.Unmarshal([]byte(variablesJSON), &variables); err != nil {
			return nil, contextureerrors.WithOpf("ParseRuleID", "invalid JSON5 variables in rule ID '%s': %w", ruleID, err)
		}
		parsed.Variables = variables
	}

	return parsed, nil
}

// FetchRule fetches a single rule from local filesystem
func (f *LocalFetcher) FetchRule(_ context.Context, ruleID string) (*domain.Rule, error) {
	log.Debug("Fetching rule from local filesystem", "ruleID", ruleID)

	// Parse the rule ID
	parsed, err := f.ParseRuleID(ruleID)
	if err != nil {
		return nil, err
	}
	log.Debug("Parsed rule ID", "originalID", ruleID, "parsedPath", parsed.RulePath, "source", parsed.Source)

	var rulePath string
	var rulesDir string

	// Check if parsed.RulePath is already an absolute path (for global local rules)
	if filepath.IsAbs(parsed.RulePath) {
		// Use the absolute path directly for global local rules
		rulePath = parsed.RulePath
		if !strings.HasSuffix(rulePath, ".md") {
			rulePath += ".md"
		}
		// Find the rules directory by looking for .contexture/rules or just rules in the path
		// For global local rules, this will be ~/.contexture/rules/
		switch {
		case strings.Contains(rulePath, filepath.Join(domain.ContextureDir, domain.LocalRulesDir)):
			// Path contains .contexture/rules
			idx := strings.Index(rulePath, filepath.Join(domain.ContextureDir, domain.LocalRulesDir))
			rulesDir = rulePath[:idx+len(filepath.Join(domain.ContextureDir, domain.LocalRulesDir))]
		case strings.Contains(rulePath, string(filepath.Separator)+domain.LocalRulesDir+string(filepath.Separator)):
			// Path contains /rules/
			idx := strings.Index(rulePath, string(filepath.Separator)+domain.LocalRulesDir+string(filepath.Separator))
			rulesDir = rulePath[:idx+len(string(filepath.Separator)+domain.LocalRulesDir)]
		default:
			// Fallback to parent directory
			rulesDir = filepath.Dir(rulePath)
		}
		log.Debug("Using absolute path for global local rule", "rulePath", rulePath, "rulesDir", rulesDir)
	} else {
		// Find the correct rules directory for project local rules
		rulesDir, err = f.findRulesDirectory()
		if err != nil {
			return nil, contextureerrors.WithOp("FetchRule", err)
		}

		// Construct full path with validation to prevent path traversal
		rulePath = filepath.Join(rulesDir, parsed.RulePath)
		if !strings.HasSuffix(rulePath, ".md") {
			rulePath += ".md"
		}

		// Validate path is within rules directory
		cleanPath, err := filepath.Abs(rulePath)
		if err != nil {
			return nil, contextureerrors.WithOp("FetchRule", err)
		}

		cleanBase, err := filepath.Abs(rulesDir)
		if err != nil {
			return nil, contextureerrors.WithOp("FetchRule", err)
		}

		// Ensure rule path is within rules directory
		if !strings.HasPrefix(cleanPath, cleanBase+string(filepath.Separator)) && cleanPath != cleanBase {
			return nil, contextureerrors.WithOpf("FetchRule", "rule path %q is outside rules directory", parsed.RulePath)
		}
	}

	// Read the file
	data, err := afero.ReadFile(f.fs, rulePath)
	if err != nil {
		return nil, contextureerrors.WithOpf("FetchRule", "failed to read rule file %s: %w", rulePath, err)
	}

	// Parse the rule - format the rule ID properly for local rules
	// For global local rules with absolute paths, extract just the relative path
	relativeRulePath := parsed.RulePath
	if filepath.IsAbs(parsed.RulePath) {
		// Extract relative path from rulesDir to rulePath (without .md extension)
		rulePathWithoutExt := strings.TrimSuffix(rulePath, ".md")
		if rel, err := filepath.Rel(rulesDir, rulePathWithoutExt); err == nil {
			relativeRulePath = rel
		}
	}

	formattedRuleID := fmt.Sprintf("[contexture(local):%s]", relativeRulePath)
	metadata := Metadata{
		ID:        formattedRuleID,
		FilePath:  relativeRulePath,
		Source:    "local",
		Ref:       "", // Local fetcher doesn't use refs
		Variables: parsed.Variables,
	}

	rule, err := f.parser.ParseRule(string(data), metadata)
	if err != nil {
		return nil, contextureerrors.WithOp("FetchRule", err)
	}

	log.Debug("Successfully fetched local rule", "ruleID", ruleID)
	return rule, nil
}

// FetchRules fetches multiple rules from local filesystem
func (f *LocalFetcher) FetchRules(ctx context.Context, ruleIDs []string) ([]*domain.Rule, error) {
	var rules []*domain.Rule
	var errors []error

	for _, ruleID := range ruleIDs {
		rule, err := f.FetchRule(ctx, ruleID)
		if err != nil {
			errors = append(errors, contextureerrors.WithOpf("FetchRules", "failed to fetch rule %s: %w", ruleID, err))
			continue
		}
		rules = append(rules, rule)
	}

	if len(errors) > 0 {
		return nil, combineErrors(errors)
	}

	return rules, nil
}

// ListAvailableRules lists all available local rules
func (f *LocalFetcher) ListAvailableRules(
	_ context.Context,
	_, _ string,
) ([]string, error) {
	// Find the correct rules directory
	rulesDir, err := f.findRulesDirectory()
	if err != nil {
		return nil, contextureerrors.WithOp("ListAvailableRules", err)
	}

	// Check if rules directory exists
	exists, err := afero.DirExists(f.fs, rulesDir)
	if err != nil {
		return nil, contextureerrors.WithOp("ListAvailableRules", err)
	}
	if !exists {
		return []string{}, nil // Return empty slice if directory doesn't exist
	}

	// For local fetcher, ignore source and branch
	var ruleFiles []string

	err = afero.Walk(f.fs, rulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			// Get relative path from rules directory
			relPath, err := filepath.Rel(rulesDir, path)
			if err != nil {
				return err
			}

			// Remove .md extension
			ruleID := strings.TrimSuffix(relPath, ".md")
			ruleFiles = append(ruleFiles, ruleID)
		}

		return nil
	})
	if err != nil {
		return nil, contextureerrors.WithOp("ListAvailableRules", err)
	}

	return ruleFiles, nil
}

// ListAvailableRulesWithStructure lists all available local rules with folder structure
func (f *LocalFetcher) ListAvailableRulesWithStructure(
	ctx context.Context,
	source, ref string,
) (*domain.RuleNode, error) {
	// Get the flat list of rules first
	ruleFiles, err := f.ListAvailableRules(ctx, source, ref)
	if err != nil {
		return nil, contextureerrors.WithOp("ListAvailableRulesWithStructure", err)
	}

	// Build the tree structure
	tree := domain.NewRuleTree(ruleFiles)
	return tree, nil
}

// findRulesDirectory discovers the correct local rules directory
func (f *LocalFetcher) findRulesDirectory() (string, error) {
	currentDir := f.baseDir
	if currentDir == "." {
		var err error
		currentDir, err = os.Getwd()
		if err != nil {
			return "", contextureerrors.WithOp("findRulesDirectory", err)
		}
	}

	// Check for config files to understand project structure
	contextureConfigPath := filepath.Join(currentDir, domain.ContextureDir, "config.yaml")
	rootConfigPath := filepath.Join(currentDir, domain.ConfigFile)
	contextureInDirConfigPath := filepath.Join(currentDir, domain.ContextureDir, domain.ConfigFile)
	contextureExists, _ := afero.Exists(f.fs, contextureConfigPath)
	rootExists, _ := afero.Exists(f.fs, rootConfigPath)
	contextureInDirExists, _ := afero.Exists(f.fs, contextureInDirConfigPath)

	// If config exists, use config-based rules directory detection
	if contextureExists || contextureInDirExists {
		// Config is in .contexture/ directory, so rules are in .contexture/rules/
		return filepath.Join(currentDir, domain.ContextureDir, domain.LocalRulesDir), nil
	} else if rootExists {
		// Config is in root, so rules are in rules/
		return filepath.Join(currentDir, domain.LocalRulesDir), nil
	}

	// No config exists - try to determine from filesystem structure
	// First check if currentDir has .md files directly (it IS the rules directory)
	files, err := afero.ReadDir(f.fs, currentDir)
	if err == nil {
		// Check for direct .md files in currentDir
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), domain.MarkdownExt) {
				return currentDir, nil
			}
		}

		// No direct .md files, check if there's a rules/ subdirectory
		rulesSubdir := filepath.Join(currentDir, domain.LocalRulesDir)
		if exists, _ := afero.DirExists(f.fs, rulesSubdir); exists {
			return rulesSubdir, nil
		}

		// No rules/ subdirectory, check if any subdirectory contains .md files
		// If so, currentDir is probably the rules directory (test scenario)
		for _, file := range files {
			if file.IsDir() {
				subDirPath := filepath.Join(currentDir, file.Name())
				if hasMarkdownFiles(f.fs, subDirPath) {
					return currentDir, nil
				}
			}
		}
	}

	// Default to rules/ subdirectory
	return filepath.Join(currentDir, domain.LocalRulesDir), nil
}

// hasMarkdownFiles checks if a directory contains .md files directly or in subdirectories
func hasMarkdownFiles(fs afero.Fs, dir string) bool {
	err := afero.Walk(fs, dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), domain.MarkdownExt) {
			return filepath.SkipAll // Found a markdown file, stop walking
		}
		return nil
	})
	return errors.Is(err, filepath.SkipAll)
}
