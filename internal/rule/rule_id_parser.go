package rule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/titanous/json5"
)

// IDParser handles parsing of rule identifiers
type IDParser interface {
	ParseRuleID(ruleID string) (*domain.ParsedRuleID, error)
}

// DefaultRuleIDParser implements rule ID parsing
type DefaultRuleIDParser struct {
	defaultURL      string
	ruleIDPattern   *regexp.Regexp
	simpleIDPattern *regexp.Regexp
}

// NewRuleIDParser creates a new rule ID parser
func NewRuleIDParser(defaultURL string) IDParser {
	return &DefaultRuleIDParser{
		defaultURL:      defaultURL,
		ruleIDPattern:   domain.RuleIDParsePatternRegex,
		simpleIDPattern: domain.SimpleRuleIDPatternRegex,
	}
}

// ParseRuleID parses a Contexture rule ID into its components
func (p *DefaultRuleIDParser) ParseRuleID(ruleID string) (*domain.ParsedRuleID, error) {
	// First try the full rule ID pattern [contexture:path] or [contexture(source):path,ref]{variables}
	matches := p.ruleIDPattern.FindStringSubmatch(ruleID)

	if len(matches) > 0 {
		parsed := &domain.ParsedRuleID{
			RulePath: matches[2], // Required path component
		}

		// Optional source (defaults to official repo)
		if matches[1] != "" {
			parsed.Source = matches[1]
		} else {
			parsed.Source = p.defaultURL
		}

		// Optional ref (branch/tag/commit, defaults to main)
		if matches[3] != "" {
			parsed.Ref = matches[3]
		} else {
			parsed.Ref = "main"
		}

		// Optional variables (JSON5 format)
		if len(matches) > 4 && matches[4] != "" {
			variables := make(map[string]any)
			if err := json5.Unmarshal([]byte(matches[4]), &variables); err != nil {
				return nil, fmt.Errorf("invalid JSON5 variables in rule ID '%s': %w", ruleID, err)
			}
			parsed.Variables = variables
		}

		return parsed, nil
	}

	// Try simple format: core/security/input-validation
	if p.simpleIDPattern.MatchString(ruleID) {
		return &domain.ParsedRuleID{
			Source:   p.defaultURL,
			RulePath: ruleID,
			Ref:      "main",
		}, nil
	}

	// Check if it's a direct Git URL
	if strings.HasPrefix(ruleID, "https://") || strings.HasPrefix(ruleID, "git@") {
		if strings.Contains(ruleID, "#") {
			parts := strings.SplitN(ruleID, "#", 2)
			return &domain.ParsedRuleID{
				Source:   parts[0],
				RulePath: parts[1],
				Ref:      "main",
			}, nil
		}
		return nil, contextureerrors.ValidationErrorf(
			"ruleID",
			"direct Git URL must include path after '#': %s",
			ruleID,
		)
	}

	return nil, contextureerrors.ValidationErrorf("ruleID", "invalid rule ID format: %s", ruleID)
}
