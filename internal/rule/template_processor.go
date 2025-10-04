package rule

import (
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/template"
)

// TemplateEngine handles template processing
type TemplateEngine interface {
	ProcessTemplate(content string, variables map[string]any) (string, error)
	ExtractVariables(template string) ([]string, error)
}

// DefaultTemplateEngine implements template processing using Go templates
type DefaultTemplateEngine struct {
	engine template.Engine
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() TemplateEngine {
	return &DefaultTemplateEngine{
		engine: template.NewEngine(),
	}
}

// ProcessTemplate processes template content with variables
func (te *DefaultTemplateEngine) ProcessTemplate(
	content string,
	variables map[string]any,
) (string, error) {
	result, err := te.engine.Render(content, variables)
	if err != nil {
		return "", contextureerrors.Wrap(err, "process template")
	}
	return result, nil
}

// ExtractVariables extracts variable names from template content
func (te *DefaultTemplateEngine) ExtractVariables(template string) ([]string, error) {
	return te.engine.ExtractVariables(template)
}
