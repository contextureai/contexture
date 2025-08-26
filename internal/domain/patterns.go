package domain

import "regexp"

// Pre-compiled regex patterns for performance
var (
	// RuleIDPatternRegex validates the complete rule ID format including optional JSON5 variables
	RuleIDPatternRegex = regexp.MustCompile(RuleIDPattern)

	// RuleIDParsePatternRegex captures components: (1) source, (2) path, (3) branch, (4) variables
	RuleIDParsePatternRegex = regexp.MustCompile(RuleIDParsePattern)

	// RuleIDExtractPatternRegex finds rule IDs in content without anchors
	RuleIDExtractPatternRegex = regexp.MustCompile(RuleIDExtractPattern)

	// SimpleRuleIDPatternRegex matches simple rule IDs
	SimpleRuleIDPatternRegex = regexp.MustCompile(SimpleRuleIDPattern)

	// FilenameCleanRegex cleans non-alphanumeric characters from filenames
	FilenameCleanRegex = regexp.MustCompile(`[^a-zA-Z0-9\-_.]`)

	// TrackingCommentRegex matches tracking comment patterns
	TrackingCommentRegex = regexp.MustCompile(RuleIDCommentPrefix + `([^-]+)` + RuleIDCommentSuffix)

	// TagValidationRegex validates tag format (alphanumeric with hyphens)
	TagValidationRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

	// VariablesPatternRegex extracts variables from rule IDs
	VariablesPatternRegex = regexp.MustCompile(`^([^{]+)(\{.*\})?\s*$`)
)
