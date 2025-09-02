---
title: Cursor IDE Technical Specification
description: A technical specification of the Cursor IDE system.
---

# Cursor IDE Technical Specification

## Overview

Cursor IDE provides a sophisticated AI-powered development environment with comprehensive rules system and MCP (Model Context Protocol) support. This specification covers the complete implementation details for rules and MCP servers in Cursor.

## Rules System

### File Format and Structure

**File Extension**: `.mdc` (Markdown Cursor format)

**File Locations**:
- **Project Rules**: `.cursor/rules/` directory at project root
- **Nested Rules**: Supported in subdirectories throughout the codebase
- **User Rules**: Not supported (project-level only)
- **Global Rules**: Not supported (project-level only)

### Rule Syntax

Cursor rules use MDC format with YAML frontmatter for configuration:

```yaml
---
description: Rule description for AI-triggered activation
globs: ["**/*.ts", "**/*.tsx"]
alwaysApply: false
---

# Rule Title

Rule content in Markdown format...

## Implementation Guidelines
- Specific instructions
- Code examples
- Best practices

@reference-file.ts  # Optional file references
```

### Supported Trigger Modes

1. **Always Apply**
   ```yaml
   ---
   alwaysApply: true
   ---
   ```
   - Rule permanently included in context
   - Use sparingly to preserve context window

2. **Manual**
   ```yaml
   ---
   alwaysApply: false
   ---
   ```
   - Activated via `@ruleName` in conversation
   - Default mode for most rules
   - Useful for specialized workflows

3. **Model Decision**
   ```yaml
   ---
   description: insert description
   alwaysApply: false
   ---
   ```
   - Include `description` field in frontmatter
   - AI determines when to include based on relevance
   - Description should be clear and specific

4. **Glob-based**
   ```yaml
   ---
   globs: "*.tsx"
   alwaysApply: false
   ---
   ```
   - Define `globs` pattern in frontmatter
   - Automatically activated when matching files are referenced
   - Example patterns: `*.test.ts`, `src/**/*.tsx`

### Rule Inclusion and Activation

- Rules apply to **Agent** and **Inline Edit** modes
- Not available in all IDE features
- Maximum 40 MCP tools can be active simultaneously
- Rules are loaded from `.cursor/rules/` directory automatically
- Nested directories supported for organization

### Writing Effective Rules

**Structure Guidelines**:
- Keep rules under 500 lines
- Use clear, actionable language
- Include concrete examples
- Reference external files with `@filename` syntax

**Best Practices**:
```markdown
---
description: TypeScript strict mode configuration
globs: ["**/*.ts", "**/*.tsx"]
---

# TypeScript Strict Mode Standards

## Configuration Requirements
Always use strict TypeScript configuration:
- `strict: true` in tsconfig.json
- No implicit `any` types
- Explicit return types for public APIs

## Implementation Pattern
```typescript
// Good: Explicit types
export function processData(input: string): ProcessedData {
  return { processed: input.toUpperCase() };
}

// Bad: Implicit any
export function processData(input) {
  return { processed: input.toUpperCase() };
}
```

@tsconfig.json  # Reference project configuration
```

## Workflows System

**Status**: ❌ Not Supported

Cursor does not implement a separate workflows system. All workflow-like functionality must be implemented through the rules system itself. Complex multi-step processes can be documented in rules but require manual execution or integration with external tools.

### Workaround Approaches

1. **Sequential Rules**: Create rules that reference each other
2. **Command Documentation**: Include shell commands in rules
3. **External Integration**: Use MCP servers for workflow automation
4. **Template Rules**: Create rules with step-by-step procedures

## MCP (Model Context Protocol) Servers

### Configuration Format

**File Format**: JSON

**Configuration Locations**:
- **Project Level**: `.cursor/mcp.json`
- **User/Global Level**: `~/.cursor/mcp.json`

### MCP Configuration Syntax

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "database": {
      "command": "node",
      "args": ["/path/to/database-server.js"],
      "env": {
        "DATABASE_URL": "postgresql://localhost/mydb"
      }
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"],
      "env": {
        "ALLOWED_DIRECTORIES": "/home/user/projects"
      }
    }
  }
}
```

### Supported Transport Protocols

1. **stdio** (Standard Input/Output)
   - Default for local servers
   - Most common implementation

2. **SSE** (Server-Sent Events)
   - For streaming responses
   - HTTP-based communication

3. **Streamable HTTP**
   - Full bidirectional communication
   - Advanced server implementations

### MCP Server Management

**GUI Configuration**:
- Access via Cursor Settings → Features → Model Context Protocol
- Visual tool for adding/removing servers
- Individual permission controls per tool

**Popular Integrations**:
- GitHub API access
- Supabase database operations
- PostgreSQL direct connections
- Filesystem operations
- Web browsing capabilities
- Slack integration

### Tool Permission Management

```json
{
  "mcpServers": {
    "example-server": {
      "command": "node",
      "args": ["server.js"],
      "permissions": {
        "tools": {
          "read_file": true,
          "write_file": false,
          "execute_command": false
        }
      }
    }
  }
}
```

## Implementation Examples

### Complete Project Setup

**Directory Structure**:
```
project/
├── .cursor/
│   ├── rules/
│   │   ├── typescript-standards.mdc
│   │   ├── react-patterns.mdc
│   │   ├── testing/
│   │   │   └── unit-test-guidelines.mdc
│   │   └── api/
│   │       └── rest-conventions.mdc
│   └── mcp.json
├── src/
└── package.json
```

### Example Rule: API Development

```yaml
---
description: REST API development standards and patterns
globs: ["**/api/**/*.ts", "**/controllers/**/*.ts"]
---

# REST API Standards

## Endpoint Structure
- Use RESTful conventions
- Implement proper status codes
- Include error handling

## Implementation Requirements
1. All endpoints must have TypeScript types
2. Use middleware for authentication
3. Implement request validation
4. Return consistent response format

## Response Format
```typescript
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
  timestamp: string;
}
```

@api-types.ts
@middleware/auth.ts
```

## Limitations and Considerations

### System Limitations
- Rules only apply to Agent and Inline Edit modes
- Maximum 40 MCP tools active simultaneously
- No user or global rules support
- No dedicated workflow system
- Context window constraints apply

### Performance Considerations
- Large rules impact context window usage
- Too many "Always Active" rules reduce effectiveness
- MCP server startup time affects first interaction
- Network latency for remote MCP servers

### Security Considerations
- Store credentials in environment variables
- Use `.env` files (not committed to version control)
- Implement minimal necessary permissions
- Regularly rotate API tokens
- Monitor MCP tool usage

## Best Practices

### Rule Organization
1. **Hierarchical Structure**: Organize rules in directories matching project structure
2. **Naming Convention**: Use descriptive, kebab-case filenames
3. **Scope Management**: Keep rules focused on single concerns
4. **Version Control**: Commit `.cursor/rules/` directory to repository

### MCP Server Setup
1. **Start Simple**: Begin with read-only operations
2. **Test Locally**: Verify server functionality before production
3. **Monitor Usage**: Track tool invocations and errors
4. **Document Configuration**: Include setup instructions in README

### Team Collaboration
1. **Shared Rules**: Standardize rules across team
2. **Review Process**: Include rules in code reviews
3. **Documentation**: Maintain rule documentation
4. **Training**: Onboard team members to rule system

### Maintenance Strategy
1. **Regular Reviews**: Audit rule effectiveness monthly
2. **Update Patterns**: Refine based on AI behavior
3. **Performance Monitoring**: Track context window usage
4. **Feedback Loop**: Gather team input on rule improvements