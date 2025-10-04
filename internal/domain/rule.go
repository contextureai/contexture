package domain

import (
	"path/filepath"
	"strings"
	"time"

	contextureerrors "github.com/contextureai/contexture/internal/errors"
)

// TriggerType represents the type of rule trigger
type TriggerType string

const (
	// TriggerAlways means the rule is always applied
	TriggerAlways TriggerType = "always"
	// TriggerManual means the rule is only applied when manually triggered
	TriggerManual TriggerType = "manual"
	// TriggerModel means the rule is applied based on AI model conditions
	TriggerModel TriggerType = "model"
	// TriggerGlob means the rule is applied when file glob patterns match
	TriggerGlob TriggerType = "glob"
)

// RuleTrigger represents the trigger configuration for a rule
type RuleTrigger struct {
	Type  TriggerType `yaml:"type"            json:"type"            validate:"required,oneof=always manual model glob"`
	Globs []string    `yaml:"globs,omitempty" json:"globs,omitempty" validate:"required_if=Type glob"`
}

// UnmarshalYAML implements custom YAML unmarshaling for RuleTrigger.
// It supports both string format ("always") and object format ({"type": "always"}).
func (rt *RuleTrigger) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First try to unmarshal as a string
	var triggerStr string
	if err := unmarshal(&triggerStr); err == nil {
		// Handle string format
		switch triggerStr {
		case "always":
			rt.Type = TriggerAlways
		case "manual":
			rt.Type = TriggerManual
		case "model":
			rt.Type = TriggerModel
		case "glob":
			rt.Type = TriggerGlob
		default:
			return contextureerrors.ValidationErrorf("trigger", "invalid trigger type: %s", triggerStr)
		}
		return nil
	}

	// Try to unmarshal as an object
	type rawRuleTrigger RuleTrigger
	var raw rawRuleTrigger
	if err := unmarshal(&raw); err != nil {
		return contextureerrors.ValidationError("trigger", err)
	}

	*rt = RuleTrigger(raw)
	return nil
}

// Rule represents a contexture rule with all its metadata and content
type Rule struct {
	// Core identification
	ID          string   `yaml:"-"           json:"id"          validate:"required"`
	Title       string   `yaml:"title"       json:"title"       validate:"required,max=80"`
	Description string   `yaml:"description" json:"description" validate:"required,max=200"`
	Tags        []string `yaml:"tags"        json:"tags"        validate:"required,min=1,max=10"`

	// Trigger configuration
	Trigger *RuleTrigger `yaml:"trigger,omitempty" json:"trigger,omitempty"`

	// Context information
	Languages  []string `yaml:"languages,omitempty"  json:"languages,omitempty"`
	Frameworks []string `yaml:"frameworks,omitempty" json:"frameworks,omitempty"`

	// Content and metadata
	Content          string         `yaml:"-"                   json:"content"             validate:"required"`
	Variables        map[string]any `yaml:"variables,omitempty" json:"variables,omitempty"`
	DefaultVariables map[string]any `yaml:"-"                   json:"defaultVariables,omitempty"`
	FilePath         string         `yaml:"-"                   json:"filePath"`
	Source           string         `yaml:"-"                   json:"source"`
	Ref              string         `yaml:"-"                   json:"ref,omitempty"`
	CreatedAt        time.Time      `yaml:"-"                   json:"createdAt,omitempty"`
	UpdatedAt        time.Time      `yaml:"-"                   json:"updatedAt,omitempty"`
}

// GetDefaultTrigger returns a default trigger for the rule if none is set
func (r *Rule) GetDefaultTrigger() *RuleTrigger {
	if r.Trigger != nil {
		return r.Trigger
	}

	return &RuleTrigger{
		Type: TriggerManual,
	}
}

// HasLanguage checks if the rule applies to a specific language
func (r *Rule) HasLanguage(language string) bool {
	for _, lang := range r.Languages {
		if strings.EqualFold(lang, language) {
			return true
		}
	}
	return false
}

// HasFramework checks if the rule applies to a specific framework
func (r *Rule) HasFramework(framework string) bool {
	for _, fw := range r.Frameworks {
		if strings.EqualFold(fw, framework) {
			return true
		}
	}
	return false
}

// HasTag checks if the rule has a specific tag
func (r *Rule) HasTag(tag string) bool {
	for _, t := range r.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// MatchesGlob checks if a file path matches any of the rule's glob patterns
func (r *Rule) MatchesGlob(filePath string) bool {
	trigger := r.GetDefaultTrigger()
	if trigger.Type != TriggerGlob {
		return false
	}

	for _, pattern := range trigger.Globs {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
	}

	return false
}

// RuleRef represents a reference to a rule in configuration
type RuleRef struct {
	ID         string         `yaml:"id"                  json:"id"`
	Source     string         `yaml:"source,omitempty"    json:"source,omitempty"`
	Ref        string         `yaml:"ref,omitempty"       json:"ref,omitempty"`
	Variables  map[string]any `yaml:"variables,omitempty" json:"variables,omitempty"`
	CommitHash string         `yaml:"commitHash"          json:"commitHash"`
	Pinned     bool           `yaml:"pinned,omitempty"    json:"pinned,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling for RuleRef.
// It parses source information from rule IDs like [contexture(local):path].
func (rr *RuleRef) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First, unmarshal into a temporary struct to avoid infinite recursion
	type rawRuleRef RuleRef
	var raw rawRuleRef
	if err := unmarshal(&raw); err != nil {
		return err
	}

	// Copy all fields
	*rr = RuleRef(raw)

	// If source is already explicitly set, don't override it
	if rr.Source != "" {
		return nil
	}

	// Parse the rule ID to extract source information
	if rr.ID != "" {
		// Check if it matches the full pattern [contexture(source):path,ref] or [contexture:path]
		matches := RuleIDParsePatternRegex.FindStringSubmatch(rr.ID)
		if len(matches) > 1 && matches[1] != "" {
			// Source was specified in rule ID like [contexture(local):path]
			rr.Source = matches[1]
		}
		// Note: We don't set a default source here - let GetSource() handle that
	}

	return nil
}

// GetSource returns the source or default to "contexture"
func (rr *RuleRef) GetSource() string {
	if rr.Source == "" {
		return "contexture"
	}
	return rr.Source
}

// GetRef returns the ref or default to "main"
func (rr *RuleRef) GetRef() string {
	if rr.Ref == "" {
		return DefaultBranch
	}
	return rr.Ref
}

// ParsedRuleID represents the components of a parsed rule identifier
type ParsedRuleID struct {
	Source    string         `json:"source,omitempty"`
	RulePath  string         `json:"rulePath"`
	Ref       string         `json:"ref,omitempty"`
	Variables map[string]any `json:"variables,omitempty"`
}

// RuleContext represents the context for rule processing
type RuleContext struct {
	Variables map[string]any `json:"variables"`
	Globals   map[string]any `json:"globals"`
}

// ProcessedRule represents a rule after template processing
type ProcessedRule struct {
	Rule        *Rule          `json:"rule"`
	Content     string         `json:"content"`
	Context     *RuleContext   `json:"context"`
	Attribution string         `json:"attribution"`
	Variables   map[string]any `json:"variables,omitempty"`
}
