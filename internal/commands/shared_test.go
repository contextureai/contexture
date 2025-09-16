package commands

import (
	"context"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//

//

func TestNewRuleGenerator(t *testing.T) {
	t.Parallel()
	fetcher := rule.NewMockFetcher(t)
	validator := rule.NewMockValidator(t)
	processor := rule.NewMockProcessor(t)
	fs := afero.NewMemMapFs()
	registry := format.NewRegistry(fs)

	generator := NewRuleGenerator(fetcher, validator, processor, registry, fs)

	assert.NotNil(t, generator)
	assert.Equal(t, fetcher, generator.ruleFetcher)
	assert.Equal(t, validator, generator.ruleValidator)
	assert.Equal(t, processor, generator.ruleProcessor)
	assert.Equal(t, registry, generator.registry)
	assert.Equal(t, fs, generator.fs)
}

func TestRuleGenerator_GenerateRules_NoRules(t *testing.T) {
	t.Parallel()
	fetcher := rule.NewMockFetcher(t)
	validator := rule.NewMockValidator(t)
	processor := rule.NewMockProcessor(t)
	fs := afero.NewMemMapFs()
	registry := format.NewRegistry(fs)

	generator := NewRuleGenerator(fetcher, validator, processor, registry, fs)

	config := &domain.Project{
		Rules: []domain.RuleRef{}, // Empty rules
	}

	formatConfigs := []domain.FormatConfig{
		{Type: domain.FormatClaude},
	}

	err := generator.GenerateRules(context.Background(), config, formatConfigs)

	// Should return nil (no error) for empty rules
	require.NoError(t, err)

	// Verify no calls were made to mocks since there are no rules
	fetcher.AssertExpectations(t)
	validator.AssertExpectations(t)
	processor.AssertExpectations(t)
}

func TestRuleGenerator_GenerateRules_NoFormats(t *testing.T) {
	t.Parallel()
	fetcher := rule.NewMockFetcher(t)
	validator := rule.NewMockValidator(t)
	processor := rule.NewMockProcessor(t)
	fs := afero.NewMemMapFs()
	registry := format.NewRegistry(fs)

	generator := NewRuleGenerator(fetcher, validator, processor, registry, fs)

	config := &domain.Project{
		Rules: []domain.RuleRef{{ID: "test/rule1"}},
	}

	formatConfigs := []domain.FormatConfig{} // Empty formats

	err := generator.GenerateRules(context.Background(), config, formatConfigs)

	// Should return error for no target formats
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no target formats available")

	// Verify no calls were made to mocks since formats are empty
	fetcher.AssertExpectations(t)
	validator.AssertExpectations(t)
	processor.AssertExpectations(t)
}
