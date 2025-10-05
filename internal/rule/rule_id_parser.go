package rule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/provider"
	"github.com/titanous/json5"
)

// IDParser handles parsing of rule identifiers
type IDParser interface {
	ParseRuleID(ruleID string) (*domain.ParsedRuleID, error)
}

// DefaultRuleIDParser implements rule ID parsing
type DefaultRuleIDParser struct {
	defaultURL       string
	providerRegistry *provider.Registry
	ruleIDPattern    *regexp.Regexp
	providerPattern  *regexp.Regexp
	simpleIDPattern  *regexp.Regexp
}

// NewRuleIDParser creates a new rule ID parser
func NewRuleIDParser(defaultURL string, providerRegistry *provider.Registry) IDParser {
	return &DefaultRuleIDParser{
		defaultURL:       defaultURL,
		providerRegistry: providerRegistry,
		ruleIDPattern:    domain.RuleIDParsePatternRegex,
		providerPattern:  domain.ProviderRuleIDPatternRegex,
		simpleIDPattern:  domain.SimpleRuleIDPatternRegex,
	}
}

// ParseRuleID parses a Contexture rule ID into its components
func (p *DefaultRuleIDParser) ParseRuleID(ruleID string) (*domain.ParsedRuleID, error) {
	// Try @provider/path format first
	if matches := p.providerPattern.FindStringSubmatch(ruleID); len(matches) > 0 {
		providerName := matches[1]
		rulePath := matches[2]

		// Resolve provider to URL
		var url string
		var err error
		if p.providerRegistry != nil {
			url, err = p.providerRegistry.Resolve(providerName)
			if err != nil {
				return nil, fmt.Errorf("unknown provider '@%s': %w", providerName, err)
			}
		} else {
			// Fallback if no registry (shouldn't happen in normal usage)
			url = p.defaultURL
		}

		return &domain.ParsedRuleID{
			Source:   url,
			RulePath: rulePath,
			Ref:      "main",
		}, nil
	}

	// Try the full rule ID pattern [contexture:path] or [contexture(source):path,ref]{variables}
	matches := p.ruleIDPattern.FindStringSubmatch(ruleID)

	if len(matches) > 0 {
		parsed := &domain.ParsedRuleID{
			RulePath: matches[2], // Required path component
		}

		// Optional source (defaults to official repo)
		source := matches[1]
		if source != "" {
			// Check if source starts with @ (provider reference)
			if strings.HasPrefix(source, "@") {
				providerName := strings.TrimPrefix(source, "@")
				if p.providerRegistry != nil {
					url, err := p.providerRegistry.Resolve(providerName)
					if err != nil {
						return nil, fmt.Errorf("unknown provider '@%s': %w", providerName, err)
					}
					parsed.Source = url
				} else {
					parsed.Source = p.defaultURL
				}
			} else {
				parsed.Source = source
			}
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
				return nil, contextureerrors.ValidationErrorf("ruleID", "invalid JSON5 variables in rule ID '%s': %v", ruleID, err)
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
