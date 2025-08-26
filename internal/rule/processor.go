package rule

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/domain"
)

// TemplateProcessor implements rule processing with separated concerns
type TemplateProcessor struct {
	templateEngine       TemplateEngine
	variableManager      VariableManager
	attributionGenerator AttributionGenerator
}

// NewProcessor creates a new processor with separated components
func NewProcessor() Processor {
	return &TemplateProcessor{
		templateEngine:       NewTemplateEngine(),
		variableManager:      NewVariableManager(),
		attributionGenerator: NewAttributionGenerator(),
	}
}

// ProcessRule processes a single rule with the given context
func (p *TemplateProcessor) ProcessRule(
	rule *domain.Rule,
	ruleContext *domain.RuleContext,
) (*domain.ProcessedRule, error) {
	log.Debug("Processing rule", "id", rule.ID)

	// Build variable map for format implementations to use
	variables := p.variableManager.BuildVariableMap(rule, ruleContext)

	// Generate attribution
	attribution := p.attributionGenerator.GenerateAttribution(rule)

	processed := &domain.ProcessedRule{
		Rule:        rule,
		Content:     rule.Content, // Pass raw content - format will handle templating
		Context:     ruleContext,
		Attribution: attribution,
		Variables:   variables, // Pass variables for format to use
	}

	log.Debug("Successfully processed rule", "id", rule.ID)
	return processed, nil
}

// ProcessRules processes multiple rules concurrently
func (p *TemplateProcessor) ProcessRules(
	rules []*domain.Rule,
	ruleContext *domain.RuleContext,
) ([]*domain.ProcessedRule, error) {
	return p.ProcessRulesWithContext(context.Background(), rules, ruleContext)
}

// ProcessRulesWithContext processes multiple rules concurrently with context cancellation
func (p *TemplateProcessor) ProcessRulesWithContext(
	ctx context.Context,
	rules []*domain.Rule,
	ruleContext *domain.RuleContext,
) ([]*domain.ProcessedRule, error) {
	if len(rules) == 0 {
		return []*domain.ProcessedRule{}, nil
	}

	log.Debug("Processing multiple rules", "count", len(rules))

	// Use a worker pool for concurrent processing
	type result struct {
		processed *domain.ProcessedRule
		err       error
		index     int
	}

	resultChan := make(chan result, len(rules))

	// Start workers
	for i, rule := range rules {
		go func(idx int, r *domain.Rule) {
			processed, err := p.ProcessRule(r, ruleContext)
			select {
			case resultChan <- result{processed: processed, err: err, index: idx}:
			case <-ctx.Done():
				// Context cancelled, exit without sending result
			}
		}(i, rule)
	}

	// Collect results in order
	results := make([]*domain.ProcessedRule, len(rules))
	var errors []error

	for range rules {
		select {
		case res := <-resultChan:
			if res.err != nil {
				errors = append(
					errors,
					fmt.Errorf("failed to process rule at index %d: %w", res.index, res.err),
				)
			} else {
				results[res.index] = res.processed
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while processing rules: %w", ctx.Err())
		}
	}

	// Return errors if any occurred
	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to process %d rules: %w", len(errors), combineErrors(errors))
	}

	// Filter out nil results
	var processed []*domain.ProcessedRule
	for _, result := range results {
		if result != nil {
			processed = append(processed, result)
		}
	}

	log.Debug("Successfully processed all rules", "count", len(processed))
	return processed, nil
}

// ProcessTemplate processes template content with variables
func (p *TemplateProcessor) ProcessTemplate(
	content string,
	variables map[string]any,
) (string, error) {
	// Enrich variables with built-ins before processing
	enrichedVars := p.variableManager.EnrichWithBuiltins(variables)
	return p.templateEngine.ProcessTemplate(content, enrichedVars)
}

// GenerateAttribution generates attribution text for a rule
func (p *TemplateProcessor) GenerateAttribution(rule *domain.Rule) string {
	return p.attributionGenerator.GenerateAttribution(rule)
}

// ValidateTemplate validates a template for syntax and variable requirements
func (p *TemplateProcessor) ValidateTemplate(templateContent string, requiredVars []string) error {
	// Check syntax by trying to process with empty variables
	_, err := p.templateEngine.ProcessTemplate(templateContent, make(map[string]any))
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}

	// If no required variables specified, we're done
	if len(requiredVars) == 0 {
		return nil
	}

	// Extract variables from template and check if all required variables are present
	templateVars, err := p.templateEngine.ExtractVariables(templateContent)
	if err != nil {
		return fmt.Errorf("failed to extract template variables: %w", err)
	}

	// Check if all required variables are present
	templateVarMap := make(map[string]bool)
	for _, v := range templateVars {
		templateVarMap[v] = true
	}

	var missingVars []string
	for _, required := range requiredVars {
		if !templateVarMap[required] {
			missingVars = append(missingVars, required)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required variables: %v", missingVars)
	}

	return nil
}
