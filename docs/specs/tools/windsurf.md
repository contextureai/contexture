# Windsurf IDE Technical Specification

## Overview

Windsurf IDE implements a comprehensive three-tier system with rules, memories, and full workflow automation capabilities. This specification details the complete implementation for rules, workflows, memories, and MCP servers in Windsurf.

## Rules System

### File Format and Structure

**File Extension**: `.md` (Standard Markdown)

**File Locations**:
- **Project Rules**: `.windsurf/rules/` directory (recommended)
- **Legacy Format**: `.windsurfrules` file in project root
- **Global Rules**: `~/.windsurf/global_rules.md`
- **User Rules**: Not separately supported (use global)

### Rule Syntax

Windsurf uses Markdown with YAML frontmatter for trigger configuration:

```markdown
---
trigger: always_on
description: Comprehensive code standards for TypeScript, React, and project conventions
---

# Development Rules

<coding_guidelines>
## Language Standards
- Programming language: TypeScript
- Use strict mode
- Prefer const over let

## Code Organization
- One component per file
- Group related functions
- Use barrel exports
</coding_guidelines>

<testing_requirements>
## Unit Testing
- Minimum 80% coverage
- Test edge cases
- Mock external dependencies
</testing_requirements>
```

### Supported Trigger Modes

1. **Always On**
   ```yaml
   ---
   trigger: always_on
   description: Comprehensive code standards for TypeScript, React, and project conventions
   ---
   ```
   - Automatically included in all conversations
   - Use sparingly due to context limits
   - Best for core project standards

2. **Manual Activation**
   ```yaml
   ---
   trigger: manual
   ---
   ```
   - Activated via `@ruleName` in Cascade chat
   - Most explicit control
   - Useful for specialized contexts

3. **Model Decision**
   ```yaml
   ---
   trigger: model_decision
   description: Insert decision criteria here
   ---
   ```
   - AI determines relevance based on context
   - Requires clear, descriptive content
   - Natural language activation

4. **File Glob Patterns**
   ```yaml
   ---
   trigger: glob
   globs: "*.ts,*.tsx"
   ---
   ```
   - Activated when matching files are referenced
   - Pattern examples: `*.test.ts`, `src/**/*.tsx`
   - Automatic contextual inclusion

### Rule Limitations

- **Individual File Limit**: 6,000 characters per rule file
- **Combined Limit**: 12,000 characters for all rules
- **Character Counting**: Includes all content and formatting

### Writing Effective Rules

**Structure Template**:
```markdown
# Rule Title

> Brief description of the rule's purpose

## Context
- **Languages**: typescript, javascript
- **Frameworks**: react, nextjs
- **Categories**: performance, security

## Guidelines

### Primary Directive
Main guidance with clear headings

### Implementation Details
Specific instructions and patterns

## Examples
```typescript
// Good pattern
const example = implementCorrectly();

// Avoid this pattern
const bad = avoidThis();
```

## References
- [Documentation Link](https://example.com)
- Related rules: @security-rules, @testing-rules
```

## Memories System

### Overview

Windsurf's unique memories system automatically generates and stores contextual information during interactions.

### Characteristics

- **Automatic Generation**: Created by AI during conversations
- **Workspace Scoped**: Stored per project workspace
- **No Credit Consumption**: Doesn't count against usage limits
- **Dynamic Evolution**: Updates based on interactions

### Memory Storage

**Location**: `.windsurf/memories/` directory

**Format**: JSON with metadata
```json
{
  "id": "mem_12345",
  "timestamp": "2024-01-15T10:30:00Z",
  "context": "Database optimization discussion",
  "content": "Project uses PostgreSQL with connection pooling",
  "confidence": 0.95,
  "references": ["file.ts", "config.json"]
}
```

### Memory Management

- Memories persist across sessions
- Can be manually edited or deleted
- Automatically deduplicated
- Ranked by relevance and recency

## Workflows System

### File Format and Structure

**Status**: ✅ Fully Supported

**File Extension**: `.md` (Markdown format)

**File Locations**:
- **Project Workflows**: `.windsurf/workflows/` directory
- **Global Workflows**: `~/.windsurf/workflows/`
- **Team Workflows**: Shared via version control

### Workflow Syntax

```markdown
# Workflow Name

## Description
Brief explanation of what this workflow accomplishes

## Prerequisites
- Required tools or dependencies
- Environment setup needed
- Access permissions required

## Parameters
- `param1`: Description of first parameter
- `param2`: Optional parameter with default

## Steps

### Step 1: Initialize
```bash
npm install
npm run setup
```

### Step 2: Build
```bash
npm run build
npm run test
```

### Step 3: Deploy
```bash
npm run deploy:staging
npm run test:e2e
npm run deploy:production
```

## Success Criteria
- All tests passing
- Deployment successful
- No console errors

## Rollback Procedure
1. Revert deployment
2. Restore database backup
3. Clear CDN cache
```

### Workflow Invocation

**Slash Commands**: `/workflow-name` in Cascade chat

**Examples**:
- `/deploy-staging`
- `/run-tests`
- `/generate-report`

### Nested Workflows

```markdown
# Main Workflow

## Steps
1. Run prerequisite checks
2. `/setup-environment` (nested workflow)
3. Execute main process
4. `/cleanup-resources` (nested workflow)
5. Generate report
```

### Workflow Limitations

- **File Size Limit**: 12,000 characters per workflow
- **Execution Time**: Subject to Cascade timeout limits
- **Concurrency**: Single workflow execution at a time

## MCP (Model Context Protocol) Servers

### Configuration Format

**File Format**: JSON (Claude Desktop-compatible schema)

**Configuration Location**:
- **Primary**: `~/.codeium/windsurf/mcp_config.json`
- **Legacy**: `~/.windsurf/mcp.json`

### MCP Configuration Syntax

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"
      },
      "transport": "stdio"
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "DATABASE_URL": "postgresql://user:pass@localhost/db"
      },
      "transport": "stdio",
      "permissions": {
        "read": true,
        "write": false
      }
    },
    "web-browser": {
      "command": "node",
      "args": ["/path/to/browser-server.js"],
      "transport": "sse",
      "whitelist": ["*.example.com", "docs.*.org"]
    }
  }
}
```

### Supported Transport Protocols

1. **stdio** (Standard Input/Output)
   - Default and most common
   - Best for local servers

2. **SSE** (Server-Sent Events)
   - Streaming responses
   - HTTP-based

### Plugin Store Integration

Windsurf includes a built-in Plugin Store for one-click MCP server installation:

**Access**: Settings → Extensions → MCP Plugin Store

**Features**:
- Pre-configured servers
- Automatic dependency installation
- Configuration templates
- Community contributions

### Enterprise Features

```json
{
  "mcpServers": {
    "corporate-api": {
      "command": "corporate-mcp-server",
      "whitelist": ["*.company.com"],
      "regex_patterns": ["^/api/v[0-9]+/.*"],
      "team_management": {
        "shared_config": true,
        "admin_override": false,
        "audit_logging": true
      },
      "rate_limiting": {
        "requests_per_minute": 60,
        "burst_size": 10
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
├── .windsurf/
│   ├── rules/
│   │   ├── coding-standards.md
│   │   ├── security-guidelines.md
│   │   └── testing-patterns.md
│   ├── workflows/
│   │   ├── deploy-production.md
│   │   ├── run-tests.md
│   │   └── update-dependencies.md
│   └── memories/
│       └── (auto-generated)
├── src/
└── package.json
```

### Example Rule: React Component Standards

```markdown
# React Component Standards

<component_guidelines>
## File Organization
- One component per file
- Colocate styles and tests
- Use index.ts for exports

## Component Structure
```tsx
import React from 'react';
import styles from './Component.module.css';

interface ComponentProps {
  // Props interface
}

export const Component: React.FC<ComponentProps> = (props) => {
  // Implementation
};
```

## State Management
- Use hooks for local state
- Context for cross-component state
- Redux for global application state
</component_guidelines>

<performance_rules>
## Optimization Requirements
- Memoize expensive computations
- Use React.memo for pure components
- Implement proper key props
- Lazy load heavy components
</performance_rules>
```

### Example Workflow: Continuous Deployment

```markdown
# Deploy to Production

## Description
Automated deployment pipeline with safety checks

## Prerequisites
- Clean git working directory
- All tests passing
- Approval from team lead

## Steps

### 1. Pre-deployment Checks
```bash
git status --porcelain
npm test
npm run lint
```

### 2. Build and Verify
```bash
npm run build:production
npm run test:integration
```

### 3. Deploy to Staging
```bash
npm run deploy:staging
npm run test:e2e:staging
```

### 4. Production Deployment
```bash
npm run backup:database
npm run deploy:production
npm run verify:production
```

### 5. Post-deployment
```bash
npm run notify:team
npm run update:documentation
```

## Rollback
```bash
npm run rollback:production
npm run restore:database
```
```

## Best Practices

### Rule Organization

1. **XML Tag Grouping**: Use tags to organize related rules
2. **Clear Headers**: Structure with markdown headers
3. **Concise Content**: Stay within character limits
4. **Version Control**: Commit `.windsurf/` directory

### Workflow Design

1. **Atomic Steps**: Each step should be independently executable
2. **Error Handling**: Include rollback procedures
3. **Documentation**: Clear descriptions and prerequisites
4. **Idempotency**: Workflows should be safely re-runnable

### Memory Optimization

1. **Review Regularly**: Check auto-generated memories
2. **Prune Outdated**: Remove irrelevant memories
3. **Supplement Manually**: Add important context
4. **Monitor Growth**: Track memory directory size

### MCP Configuration

1. **Security First**: Use environment variables for secrets
2. **Minimal Permissions**: Start with read-only access
3. **Test Thoroughly**: Verify server functionality
4. **Document Setup**: Include configuration instructions

## Limitations and Considerations

### System Limitations

- **Rules**: 6,000 char/file, 12,000 total
- **Workflows**: 12,000 characters per workflow
- **Memories**: Automatic generation only
- **MCP**: Manual approval required for tools

### Performance Considerations

- Character limits affect complex rules
- Workflow execution timeout constraints
- Memory accumulation over time
- MCP server startup overhead

### Security Considerations

- Credentials in environment variables
- Whitelist patterns for MCP servers
- Audit logging for enterprise features
- Regular security reviews-

## Team Collaboration

### Shared Configuration

1. **Rules**: Version control `.windsurf/rules/`
2. **Workflows**: Share via repository
3. **MCP Config**: Use template files
4. **Documentation**: Maintain team wiki

### Best Practices

1. **Naming Conventions**: Consistent file naming
2. **Review Process**: Include in PRs
3. **Testing**: Validate workflows before sharing
4. **Training**: Team onboarding sessions