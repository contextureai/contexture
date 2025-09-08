package domain

// Configuration file names
const (
	ConfigFile    = ".contexture.yaml"
	ContextureDir = ".contexture"
	LocalRulesDir = "rules"
	TemplateFile  = "CLAUDE_TEMPLATE.md"
)

// Output file defaults
const (
	ClaudeOutputFile   = "CLAUDE.md"
	CursorOutputDir    = ".cursor/rules"
	WindsurfOutputDir  = ".windsurf/rules"
	WindsurfOutputFile = ".windsurfrules"
)

// Default repository configuration
const (
	DefaultRepository = "https://github.com/contextureai/rules.git"
	DefaultBranch     = "main"
	DefaultSource     = "contexture"
)

// Rule identifier patterns
const (
	// RuleIDPattern validates the complete rule ID format including optional JSON5 variables
	RuleIDPattern = `^\[contexture(?:\([^)]+\))?:[^]]+(?:,[^]]+)?\](?:\s*\{.*\})?\s*$`
	// RuleIDParsePattern captures components: (1) source, (2) path, (3) branch, (4) variables
	RuleIDParsePattern = `\[contexture(?:\(([^)]+)\))?:([^,\]]+)(?:,([^]]+))?\](?:\s*(\{.*\}))?\s*$`
	// RuleIDExtractPattern finds rule IDs in content without anchors
	RuleIDExtractPattern = `\[contexture(?:\([^)]+\))?:[^]]+(?:,[^]]+)?\](?:\s*\{[^}]*\})?`
	SimpleRuleIDPattern  = `^[a-zA-Z0-9_/-]+(?:\s*\{.*\})?\s*$`
)

// File extensions
const (
	MarkdownExt = ".md"
	CursorExt   = ".mdc"
	YAMLExt     = ".yaml"
	YMLExt      = ".yml"
)

// Git configuration
const (
	GitTimeout = 30 // seconds
	MaxDepth   = 1  // For shallow clones
)

// Template placeholders
const (
	RulesContentPlaceholder = "{RULES_CONTENT}"
	RuleIDCommentPrefix     = "<!-- id: "
	RuleIDCommentSuffix     = " -->"
)

// Validation limits
const (
	MaxTitleLength         = 80
	MaxDescriptionLength   = 200
	MaxTags                = 10
	MinTags                = 1
	MaxParallelFetches     = 20
	DefaultParallelFetches = 5
	MaxRuleIDLength        = 200
	MaxURLLength           = 500
)

// Format-specific limits
const (
	// Windsurf character limits per specification
	WindsurfMaxSingleRuleChars = 12000 // Maximum characters per individual rule file
)

// Rule fetching configuration
const (
	DefaultFetchTimeout = 10 * 60 // seconds (10 minutes)
	DefaultMaxWorkers   = 5
)

// File permissions
const (
	DirPermission  = 0o750 // Directory permissions
	FilePermission = 0o644 // File permissions
)

// UI Configuration
const (
	DefaultTerminalWidth = 80
	ProgressSpinnerDelay = 100 // milliseconds
)

// Error codes for consistent error handling
const (
	ExitSuccess      = 0
	ExitError        = 1
	ExitUsageError   = 2
	ExitConfigError  = 3
	ExitNetworkError = 4
	ExitNotFound     = 5
)
