// Package fixtures provides test data for end-to-end tests
package fixtures

// SampleRules contains predefined rule content for testing
var SampleRules = map[string]string{
	"security/input-validation": `---
title: "Input Validation"
description: "Validate and sanitize all user inputs to prevent security vulnerabilities"
tags: ["security", "validation", "xss", "injection"]
languages: ["javascript", "typescript", "python", "go"]
trigger:
  type: glob
  globs: ["*.js", "*.ts", "*.py", "*.go"]
---

# Input Validation

## Overview
Always validate and sanitize user inputs to prevent security vulnerabilities including XSS, SQL injection, and command injection attacks.

## Implementation
- Use parameterized queries for database operations
- Validate input types, lengths, and formats
- Sanitize output when displaying user content
- Use allowlist validation when possible

## Examples

### JavaScript/TypeScript
` + "```javascript" + `
function validateEmail(email) {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}
` + "```" + `

### Python
` + "```python" + `
import re

def validate_email(email):
    pattern = r'^[^\s@]+@[^\s@]+\.[^\s@]+$'
    return re.match(pattern, email) is not None
` + "```" + ``,

	"performance/caching": `---
title: "Caching Strategy"
description: "Implement effective caching to improve application performance"
tags: ["performance", "caching", "optimization"]
languages: ["javascript", "typescript", "python", "go", "java"]
---

# Caching Strategy

## Overview
Implement multi-layer caching strategy to reduce database load and improve response times.

## Cache Levels
1. **Browser Cache** - Static assets and API responses
2. **CDN Cache** - Global content distribution
3. **Application Cache** - In-memory caching of frequently accessed data
4. **Database Cache** - Query result caching

## Implementation Guidelines
- Set appropriate cache TTL values
- Implement cache invalidation strategies
- Use cache-aside pattern for dynamic data
- Monitor cache hit ratios

## Best Practices
- Cache at multiple levels
- Implement graceful cache failures
- Use cache warming for critical data
- Monitor and tune cache performance`,

	"testing/unit-tests": `---
title: "Unit Testing Standards"
description: "Comprehensive unit testing guidelines for reliable code"
tags: ["testing", "unit-tests", "quality-assurance"]
languages: ["javascript", "typescript", "python", "go", "java"]
frameworks: ["jest", "pytest", "go-test", "junit"]
---

# Unit Testing Standards

## Overview
Write comprehensive unit tests to ensure code reliability and catch regressions early.

## Testing Principles
- **Arrange, Act, Assert** pattern
- Test one thing at a time
- Use descriptive test names
- Mock external dependencies

## Coverage Goals
- Minimum 80% code coverage
- 100% coverage for critical business logic
- Test both happy path and error scenarios

## Best Practices
- Write tests before or alongside code (TDD/BDD)
- Keep tests fast and independent
- Use proper test data management
- Regularly review and update tests`,

	"docs/api-documentation": `---
title: "API Documentation Standards"
description: "Maintain comprehensive and up-to-date API documentation"
tags: ["documentation", "api", "openapi", "swagger"]
frameworks: ["openapi", "swagger", "postman"]
---

# API Documentation Standards

## Overview
Maintain comprehensive, accurate, and up-to-date API documentation for all endpoints.

## Documentation Requirements
- Complete endpoint descriptions
- Request/response examples
- Parameter definitions
- Error code explanations
- Authentication requirements

## Tools and Standards
- Use OpenAPI/Swagger specification
- Generate documentation from code annotations
- Provide interactive API explorers
- Version documentation alongside API changes

## Maintenance
- Update docs with every API change
- Review documentation in code reviews
- Test examples and ensure they work
- Gather feedback from API consumers`,

	"patterns/error-handling": `---
title: "Error Handling Patterns"
description: "Consistent error handling and reporting patterns"
tags: ["error-handling", "patterns", "reliability"]
languages: ["javascript", "typescript", "python", "go", "java"]
---

# Error Handling Patterns

## Overview
Implement consistent error handling patterns across the application for better reliability and debugging.

## Error Handling Strategy
1. **Fail Fast** - Detect errors early and fail immediately
2. **Graceful Degradation** - Provide fallback functionality when possible
3. **Comprehensive Logging** - Log errors with context for debugging
4. **User-Friendly Messages** - Show helpful error messages to users

## Implementation Patterns

### Structured Error Types
- Define custom error types for different scenarios
- Include error codes and categorization
- Provide context and suggestions for resolution

### Error Boundaries
- Implement error boundaries at appropriate levels
- Prevent cascading failures
- Provide recovery mechanisms where possible

## Best Practices
- Always handle errors explicitly
- Log errors with sufficient context
- Provide actionable error messages
- Test error scenarios thoroughly`,
}

// SampleConfigs contains predefined configuration files for testing
var SampleConfigs = map[string]string{
	"basic": `version: "1.0"
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
rules: []
`,

	"all-formats": `version: "1.0"
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
  - type: windsurf
    enabled: true
rules: []
`,

	"claude-only": `version: "1.0"
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: false
  - type: windsurf
    enabled: false
rules: []
`,

	"with-remote-rules": `version: "1.0"
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
rules:
  - id: "[contexture:security/input-validation]"
  - id: "[contexture:performance/caching]"
  - id: "[contexture:testing/unit-tests]"
`,

	"contexture-dir": `version: "1.0"
formats:
  - type: claude
    enabled: true
    config:
      outputPath: CLAUDE.md
  - type: cursor
    enabled: true
    config:
      outputPath: .cursor/rules
rules: []
`,
}

// ExpectedOutputs contains expected content for output files
var ExpectedOutputs = map[string]string{
	"claude-header": `# Contexture Rules

This file contains AI assistant rules generated by Contexture.`,

	"claude-rule-section": `## Input Validation

Validate and sanitize all user inputs to prevent security vulnerabilities`,

	"cursor-structure": `.cursor/
└── rules/`,
}

// GetSampleRule returns a sample rule by name
func GetSampleRule(name string) string {
	if rule, exists := SampleRules[name]; exists {
		return rule
	}
	return ""
}

// GetSampleConfig returns a sample config by name
func GetSampleConfig(name string) string {
	if config, exists := SampleConfigs[name]; exists {
		return config
	}
	return ""
}

// GetAllRuleNames returns all available sample rule names
func GetAllRuleNames() []string {
	names := make([]string, 0, len(SampleRules))
	for name := range SampleRules {
		names = append(names, name)
	}
	return names
}
