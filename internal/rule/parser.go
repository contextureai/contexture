package rule

import (
	"maps"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/validation"
	"gopkg.in/yaml.v3"
)

// YAMLParser provides a cleaner implementation of the parser
type YAMLParser struct {
	validator validation.Validator
}

// FailsafeParser is a fallback parser that fails gracefully
type FailsafeParser struct {
	err error
}

// NewParser creates a new parser
func NewParser() Parser {
	v, err := validation.NewValidator()
	if err != nil {
		// Log the error and return a parser that always fails gracefully
		log.Error("Failed to create validator for parser, using failsafe", "error", err)
		return &FailsafeParser{err: err}
	}
	return &YAMLParser{
		validator: v,
	}
}

// ParseRule parses a rule from content with metadata
func (p *YAMLParser) ParseRule(content string, metadata Metadata) (*domain.Rule, error) {
	// Parse frontmatter and body
	frontmatter, body, err := p.ParseContent(content)
	if err != nil {
		return nil, contextureerrors.Wrap(err, "parse rule content")
	}

	// Create rule with metadata
	rule := &domain.Rule{
		ID:        metadata.ID,
		FilePath:  metadata.FilePath,
		Source:    metadata.Source,
		Ref:       metadata.Ref,
		Content:   body,
		Variables: metadata.Variables,
	}

	// Use struct-based frontmatter parsing
	fm := &ruleFrontmatter{}
	if err := p.unmarshalFrontmatter(frontmatter, fm); err != nil {
		return nil, contextureerrors.Wrap(err, "parse frontmatter")
	}

	// Map frontmatter to rule
	p.mapFrontmatterToRule(fm, rule)

	// Validate rule
	if err := p.ValidateRule(rule); err != nil {
		return nil, err
	}

	return rule, nil
}

// ruleFrontmatter represents the expected frontmatter structure
type ruleFrontmatter struct {
	Title       string              `yaml:"title"`
	Description string              `yaml:"description"`
	Tags        []string            `yaml:"tags"`
	Trigger     *domain.RuleTrigger `yaml:"trigger,omitempty"`
	Languages   []string            `yaml:"languages,omitempty"`
	Frameworks  []string            `yaml:"frameworks,omitempty"`
	Variables   map[string]any      `yaml:"variables,omitempty"`
}

// ParseContent parses frontmatter and body from content
func (p *YAMLParser) ParseContent(content string) (map[string]any, string, error) {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Check for frontmatter
	if !strings.HasPrefix(content, "---") {
		return nil, content, nil
	}

	// Remove leading ---
	content = strings.TrimPrefix(content, "---\n")

	// Find end of frontmatter
	parts := strings.SplitN(content, "\n---\n", 2)
	if len(parts) != 2 {
		return nil, "", contextureerrors.WithOpf("parse content", "invalid frontmatter format: missing closing ---")
	}

	// Parse YAML frontmatter
	var frontmatter map[string]any
	if err := yaml.Unmarshal([]byte(parts[0]), &frontmatter); err != nil {
		return nil, "", contextureerrors.Wrap(err, "parse YAML frontmatter")
	}

	// Return frontmatter and body
	body := parts[1]
	return frontmatter, body, nil
}

// ValidateRule validates a rule
func (p *YAMLParser) ValidateRule(rule *domain.Rule) error {
	result := p.validator.ValidateRule(rule)
	if !result.Valid {
		var errMsgs []string
		for _, err := range result.Errors {
			errMsgs = append(errMsgs, err.Error())
		}
		return contextureerrors.WithOpf("validate rule", "validation errors: %s", strings.Join(errMsgs, ", "))
	}
	return nil
}

// FailsafeParser methods - all return errors due to initialization failure

// ParseRule returns an error for FailsafeParser
func (f *FailsafeParser) ParseRule(_ string, _ Metadata) (*domain.Rule, error) {
	return nil, contextureerrors.Wrap(f.err, "parser initialization failed")
}

// ParseContent returns an error for FailsafeParser
func (f *FailsafeParser) ParseContent(_ string) (map[string]any, string, error) {
	return nil, "", contextureerrors.Wrap(f.err, "parser initialization failed")
}

// ValidateRule returns an error for FailsafeParser
func (f *FailsafeParser) ValidateRule(_ *domain.Rule) error {
	return contextureerrors.Wrap(f.err, "parser initialization failed")
}

// unmarshalFrontmatter unmarshals frontmatter into a struct
func (p *YAMLParser) unmarshalFrontmatter(
	data map[string]any,
	fm *ruleFrontmatter,
) error {
	// Convert map to YAML bytes then unmarshal to struct
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return contextureerrors.Wrap(err, "marshal frontmatter")
	}

	if err := yaml.Unmarshal(yamlBytes, fm); err != nil {
		return contextureerrors.Wrap(err, "unmarshal frontmatter")
	}

	return nil
}

// mapFrontmatterToRule maps frontmatter struct to rule
func (p *YAMLParser) mapFrontmatterToRule(fm *ruleFrontmatter, rule *domain.Rule) {
	rule.Title = fm.Title
	rule.Description = fm.Description
	rule.Tags = fm.Tags
	rule.Trigger = fm.Trigger
	rule.Languages = fm.Languages
	rule.Frameworks = fm.Frameworks

	// Store default variables from frontmatter
	if fm.Variables != nil {
		rule.DefaultVariables = make(map[string]any)
		maps.Copy(rule.DefaultVariables, fm.Variables)
	}

	// Merge variables - frontmatter takes precedence
	if fm.Variables != nil {
		if rule.Variables == nil {
			rule.Variables = make(map[string]any)
		}
		maps.Copy(rule.Variables, fm.Variables)
	}
}

// parseTrigger parses trigger configuration from frontmatter
func (p *YAMLParser) parseTrigger(trigger any) (*domain.RuleTrigger, error) {
	if trigger == nil {
		// No trigger configured - this is valid
		return (*domain.RuleTrigger)(nil), nil
	}

	// Handle string format
	if triggerStr, ok := trigger.(string); ok {
		ruleTrigger := &domain.RuleTrigger{}
		switch triggerStr {
		case "glob":
			ruleTrigger.Type = domain.TriggerGlob
		case "always":
			ruleTrigger.Type = domain.TriggerAlways
		case "manual":
			ruleTrigger.Type = domain.TriggerManual
		case "model":
			ruleTrigger.Type = domain.TriggerModel
		default:
			return nil, contextureerrors.WithOpf("parse trigger", "invalid trigger type: %s", triggerStr)
		}
		return ruleTrigger, nil
	}

	// Handle object format
	triggerMap, ok := trigger.(map[string]any)
	if !ok {
		return nil, contextureerrors.WithOpf("parse trigger", "trigger must be string or object")
	}

	ruleTrigger := &domain.RuleTrigger{}

	// Parse type
	if triggerType, ok := triggerMap["type"].(string); ok {
		switch triggerType {
		case "glob":
			ruleTrigger.Type = domain.TriggerGlob
		case "always":
			ruleTrigger.Type = domain.TriggerAlways
		case "manual":
			ruleTrigger.Type = domain.TriggerManual
		case "model":
			ruleTrigger.Type = domain.TriggerModel
		default:
			return nil, contextureerrors.WithOpf("parse trigger", "invalid trigger type: %s", triggerType)
		}
	} else {
		return nil, contextureerrors.WithOpf("parse trigger", "trigger type is required")
	}

	// Parse globs for glob trigger
	if ruleTrigger.Type == domain.TriggerGlob {
		if globs, ok := triggerMap["globs"]; ok {
			globList, err := p.parseStringSlice(globs, "globs")
			if err != nil {
				return nil, err
			}
			ruleTrigger.Globs = globList
		} else {
			return nil, contextureerrors.WithOpf("parse trigger", "glob trigger requires globs field")
		}
	}

	return ruleTrigger, nil
}

// parseStringSlice converts various types to []string
func (p *YAMLParser) parseStringSlice(value any, fieldName string) ([]string, error) {
	switch v := value.(type) {
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				result[i] = s
			} else {
				return nil, contextureerrors.WithOpf("parse string slice", "%s[%d] must be string", fieldName, i)
			}
		}
		return result, nil
	case []string:
		return v, nil
	case string:
		// Single string -> slice with one element
		return []string{v}, nil
	default:
		return nil, contextureerrors.WithOpf("parse string slice", "%s must be string or array of strings", fieldName)
	}
}
