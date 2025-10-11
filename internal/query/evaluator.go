// Package query provides rule query and filtering functionality
package query

import (
	"strings"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Evaluator provides methods for evaluating queries against rules
type Evaluator interface {
	// MatchesText performs simple text search in rule title and ID
	MatchesText(rule *domain.Rule, query string) bool

	// EvaluateExpr evaluates an expr expression against a rule
	EvaluateExpr(rule *domain.Rule, exprStr string) (bool, error)
}

// evaluator implements the Evaluator interface
type evaluator struct {
	// Cache compiled programs for performance
	programCache map[string]*vm.Program
}

// NewEvaluator creates a new query evaluator
func NewEvaluator() Evaluator {
	return &evaluator{
		programCache: make(map[string]*vm.Program),
	}
}

// MatchesText performs simple text search in rule title and ID
// All terms must match (AND logic) in either the title or ID
func (e *evaluator) MatchesText(rule *domain.Rule, query string) bool {
	if query == "" {
		return true
	}

	// Split query into terms
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return true
	}

	// Create searchable text from title and ID
	searchText := strings.ToLower(rule.Title + " " + rule.ID)

	// All terms must be present (AND logic)
	for _, term := range terms {
		if !strings.Contains(searchText, term) {
			return false
		}
	}

	return true
}

// EvaluateExpr evaluates an expr expression against a rule
func (e *evaluator) EvaluateExpr(rule *domain.Rule, exprStr string) (bool, error) {
	if exprStr == "" {
		return true, nil
	}

	// Check cache for compiled program
	program, ok := e.programCache[exprStr]
	if !ok {
		// Compile the expression
		compiled, err := expr.Compile(exprStr, expr.Env(buildExprEnv(nil)))
		if err != nil {
			return false, contextureerrors.ValidationError("expression", err)
		}
		program = compiled
		e.programCache[exprStr] = program
	}

	// Build environment for this rule
	env := buildExprEnv(rule)

	// Execute the program
	output, err := expr.Run(program, env)
	if err != nil {
		return false, contextureerrors.Wrap(err, "evaluate expression")
	}

	// Convert output to boolean
	result, ok := output.(bool)
	if !ok {
		return false, contextureerrors.ValidationErrorf("expression",
			"expression must return a boolean value, got %T", output)
	}

	return result, nil
}

// buildExprEnv creates an environment map for expr evaluation
func buildExprEnv(rule *domain.Rule) map[string]interface{} {
	if rule == nil {
		// Return empty environment for type checking during compilation
		return map[string]interface{}{
			"ID":          "",
			"Title":       "",
			"Description": "",
			"Tags":        []string{},
			"Languages":   []string{},
			"Frameworks":  []string{},
			"Content":     "",
			"Tag":         "",
			"Language":    "",
			"Framework":   "",
			"Provider":    "",
			"Path":        "",
			"HasVars":     false,
			"VarCount":    0,
			"Variables":   map[string]interface{}{},
			"TriggerType": "",
			"Source":      "",
			"FilePath":    "",
		}
	}

	// Extract provider and path
	provider := rule.Source
	if provider == "" {
		provider = domain.DefaultProviderName
	}

	path := domain.ExtractRulePath(rule.ID)

	// Determine trigger type
	triggerType := "manual"
	if rule.Trigger != nil {
		triggerType = string(rule.Trigger.Type)
	}

	return map[string]interface{}{
		// Direct fields
		"ID":          rule.ID,
		"Title":       rule.Title,
		"Description": rule.Description,
		"Tags":        rule.Tags,
		"Languages":   rule.Languages,
		"Frameworks":  rule.Frameworks,
		"Content":     rule.Content,
		"Variables":   rule.Variables,
		"Source":      rule.Source,
		"FilePath":    rule.FilePath,

		// Computed/convenience fields
		"Tag":       strings.Join(rule.Tags, " "),
		"Language":  strings.Join(rule.Languages, " "),
		"Framework": strings.Join(rule.Frameworks, " "),
		"Provider":  provider,
		"Path":      path,
		"HasVars":   len(rule.Variables) > 0,
		"VarCount":  len(rule.Variables),

		// Trigger info
		"TriggerType": triggerType,
	}
}
