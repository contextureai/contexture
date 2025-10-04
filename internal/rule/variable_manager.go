package rule

import (
	"time"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/version"
)

// VariableManager defines the interface for managing template variables with precedence.
// It provides methods to build complete variable maps for template processing and enrich
// them with built-in system variables. Variable precedence from lowest to highest:
// globals < context variables < rule variables.
type VariableManager interface {
	BuildVariableMap(rule *domain.Rule, context *domain.RuleContext) map[string]any
	EnrichWithBuiltins(variables map[string]any) map[string]any
}

// DefaultVariableManager is the default implementation of VariableManager.
// It handles variable merging with proper precedence ordering and enriches
// variable maps with built-in variables including date/time and contexture metadata.
type DefaultVariableManager struct{}

// NewVariableManager creates a new DefaultVariableManager instance.
// The returned VariableManager can be used to build variable maps for template
// processing with automatic precedence handling and built-in variable enrichment.
func NewVariableManager() VariableManager {
	return &DefaultVariableManager{}
}

// BuildVariableMap constructs a complete variable map for template processing with proper precedence.
// Variables are merged in order of increasing precedence: globals < context variables < rule variables.
// The resulting map also includes built-in variables (date, time, contexture metadata) and a "rule"
// object containing the rule's metadata fields for template access.
//
// Parameters:
//   - rule: The rule containing rule-specific variables (highest precedence)
//   - context: The rule context containing globals and context-specific variables
//
// Returns a map suitable for use with Go's text/template engine.
func (vm *DefaultVariableManager) BuildVariableMap(
	rule *domain.Rule,
	context *domain.RuleContext,
) map[string]any {
	variables := make(map[string]any)

	// Add globals first (lowest precedence)
	if context != nil && context.Globals != nil {
		for k, v := range context.Globals {
			variables[k] = v
		}
	}

	// Add context variables (override globals)
	if context != nil && context.Variables != nil {
		for k, v := range context.Variables {
			variables[k] = v
		}
	}

	// Add rule variables last (highest precedence - override everything)
	if rule.Variables != nil {
		for k, v := range rule.Variables {
			variables[k] = v
		}
	}

	// Add the rule object as a map for template access
	variables["rule"] = vm.buildRuleMap(rule)

	// Add built-in variables
	variables = vm.addBuiltinVariables(variables)

	return variables
}

// EnrichWithBuiltins adds built-in variables to an existing variable map.
// Built-in variables include date/time helpers (now, date, time, datetime, timestamp, year)
// and contexture metadata (version, engine, build information). This method creates a new
// map with all existing variables plus the built-ins, without modifying the input.
//
// Parameters:
//   - variables: The existing variable map to enrich
//
// Returns a new map containing all original variables plus built-in variables.
func (vm *DefaultVariableManager) EnrichWithBuiltins(variables map[string]any) map[string]any {
	enriched := make(map[string]any)

	// Copy existing variables
	for k, v := range variables {
		enriched[k] = v
	}

	// Add date/time helpers
	now := time.Now()
	enriched["now"] = now.Format("2006-01-02 15:04:05")
	enriched["date"] = now.Format("2006-01-02")
	enriched["time"] = now.Format("15:04:05")
	enriched["datetime"] = now.Format("2006-01-02 15:04:05")
	enriched["timestamp"] = now.Unix()
	enriched["year"] = now.Year()

	// Add contexture-specific variables if not already present
	if _, exists := enriched["contexture"]; !exists {
		buildInfo := version.Get()
		enriched["contexture"] = map[string]any{
			"version": buildInfo.Version,
			"engine":  "go",
			"build": map[string]any{
				"version":   buildInfo.Version,
				"commit":    buildInfo.Commit,
				"date":      buildInfo.BuildDate,
				"by":        buildInfo.BuildBy,
				"goVersion": buildInfo.GoVersion,
				"platform":  buildInfo.Platform,
			},
		}
	}

	return enriched
}

// buildRuleMap converts a Rule struct to a map for template access
func (vm *DefaultVariableManager) buildRuleMap(rule *domain.Rule) map[string]any {
	ruleMap := map[string]any{
		"id":          rule.ID,
		"title":       rule.Title,
		"description": rule.Description,
		"tags":        rule.Tags,
		"languages":   rule.Languages,
		"frameworks":  rule.Frameworks,
		"source":      rule.Source,
		"ref":         rule.Ref,
		"filepath":    rule.FilePath,
	}

	// Add trigger information if present
	if rule.Trigger != nil {
		ruleMap["trigger"] = map[string]any{
			"type":  string(rule.Trigger.Type),
			"globs": rule.Trigger.Globs,
		}
	}

	return ruleMap
}

// addBuiltinVariables adds built-in variables to the map
func (vm *DefaultVariableManager) addBuiltinVariables(variables map[string]any) map[string]any {
	if variables == nil {
		variables = make(map[string]any)
	}

	// Add contexture metadata
	contextureInfo := make(map[string]any)
	contextureInfo["version"] = version.GetShort()
	contextureInfo["engine"] = "go"

	// Add detailed build information
	buildInfo := version.Get()
	contextureInfo["build"] = map[string]any{
		"version":   buildInfo.Version,
		"commit":    buildInfo.Commit,
		"date":      buildInfo.BuildDate,
		"by":        buildInfo.BuildBy,
		"goVersion": buildInfo.GoVersion,
		"platform":  buildInfo.Platform,
	}

	variables["contexture"] = contextureInfo

	// Add date/time helpers
	now := time.Now()
	variables["now"] = now
	variables["date"] = now.Format("2006-01-02")
	variables["time"] = now.Format("15:04:05")
	variables["datetime"] = now.Format("2006-01-02 15:04:05")
	variables["timestamp"] = now.Unix()

	return variables
}
