package rule

import (
	"time"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/version"
)

// VariableManager handles variable management for rule processing
type VariableManager interface {
	BuildVariableMap(rule *domain.Rule, context *domain.RuleContext) map[string]any
	EnrichWithBuiltins(variables map[string]any) map[string]any
}

// DefaultVariableManager implements variable management
type DefaultVariableManager struct{}

// NewVariableManager creates a new variable manager
func NewVariableManager() VariableManager {
	return &DefaultVariableManager{}
}

// BuildVariableMap creates a complete variable map for template processing
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

// EnrichWithBuiltins adds built-in variables to an existing variable map
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
