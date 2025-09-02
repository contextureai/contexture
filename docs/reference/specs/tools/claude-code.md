---
title: Claude Code Technical Specification
description: A technical specification of the Claude Code system.
---

# Claude Code Technical Specification

## Overview

Claude Code employs a distinctive CLAUDE.md "constitution" system with immutable system-level authority, command-based workflows, and comprehensive MCP integration. This specification details the complete implementation for the CLAUDE.md system, workflows, and MCP servers in Claude Code.

## Rules System (CLAUDE.md)

### File Format and Structure

**File Name**: `CLAUDE.md` (case-sensitive, required naming)

**File Extension**: `.md` (Standard Markdown)

**File Locations** (in precedence order):
1. **Local Override**: `CLAUDE.local.md` (highest priority)
2. **Project Root**: `CLAUDE.md` in current project
3. **Child Directories**: Loaded on-demand when navigating
4. **Parent Directories**: Inherited from parent folders
5. **Home Directory**: `~/.claude/CLAUDE.md` (global defaults)

### CLAUDE.md Syntax

The CLAUDE.md file uses XML tags for system-level instructions:

```markdown
# Project Configuration

<system-reminder>
You are working on an e-commerce platform built with Next.js 14 App Router.

## Core Principles
- Always use TypeScript strict mode
- Implement error boundaries for all components
- Use server components by default
- Client components only when necessary

## Project Structure
- /app - Next.js app router pages
- /components - Reusable React components  
- /lib - Utility functions and helpers
- /api - API route handlers

## Key Commands
- `npm run dev` - Start development server
- `npm run test` - Run test suite
- `npm run build` - Production build
- `npm run deploy` - Deploy to Vercel
</system-reminder>

# Additional Context

## Architecture Decisions
- Database: PostgreSQL with Prisma ORM
- Authentication: NextAuth.js
- Styling: Tailwind CSS
- State Management: Zustand for client state
```

### Constitution System Characteristics

1. **Immutable Authority**: CLAUDE.md content treated as system-level truth
2. **Every Interaction**: Included in all conversations automatically
3. **Hierarchical Loading**: Parent directories provide defaults
4. **Override Capability**: Local files override inherited settings
5. **XML Tag Requirement**: Content must be in `<system-reminder>` tags

### Writing Effective CLAUDE.md Files

**Structure Template**:
```markdown
# Project Name

<system-reminder>
## Project Overview
Brief description of the project and its purpose

## Tech Stack
- Framework: Next.js 14
- Language: TypeScript
- Database: PostgreSQL
- Styling: Tailwind CSS

## Project Structure
/src
  /app - Application routes
  /components - React components
  /lib - Utilities
  /types - TypeScript definitions

## Development Workflow
1. Write tests first (TDD)
2. Implement features
3. Run linting and formatting
4. Commit with conventional commits

## Code Standards
- Use functional components
- Implement proper error handling
- Write comprehensive tests
- Document complex logic

## Common Commands
```bash
npm run dev        # Development server
npm run test       # Run tests
npm run lint       # Check code quality
npm run build      # Production build
```

## File Boundaries
- Safe to modify: /src, /tests
- Never modify: /node_modules, /.next, /dist
- Ask before changing: /public, configuration files

## Important Notes
- API keys in .env.local (never commit)
- Database migrations require review
- Deploy only from main branch
</system-reminder>
```

### CLAUDE.md Best Practices

1. **Conciseness**: Keep instructions brief and clear
2. **Specificity**: Avoid generic programming advice
3. **Context Window**: Consider impact on token usage
4. **Project-Specific**: Focus on unique project requirements
5. **Command Documentation**: Include frequently used commands

## Workflows System

### Status: ✅ Fully Supported (Multiple Mechanisms)

Claude Code implements workflows through several interconnected systems:

### 1. Custom Slash Commands

**Location**: `.claude/commands/` directory

**File Format**: Markdown files with command definitions

```markdown
# /deploy Command

## Description
Deploy application to production environment

## Steps
1. Run test suite
2. Build production bundle
3. Deploy to hosting platform
4. Run smoke tests
5. Notify team

## Implementation
```bash
#!/bin/bash
npm test && \
npm run build && \
npm run deploy:prod && \
npm run test:smoke && \
npm run notify:deployment
```

## Parameters
- --env: Target environment (staging/production)
- --skip-tests: Skip test execution
- --dry-run: Simulate deployment
```

**Invocation**: Type `/deploy` in Claude Code terminal

### 2. Hooks System

**Location**: `.claude/hooks/` directory

**Lifecycle Hooks**:
- `pre-commit.md`: Before git commits
- `post-commit.md`: After git commits
- `pre-push.md`: Before git push
- `on-save.md`: On file save
- `on-start.md`: When Claude Code starts

**Example Hook** (`pre-commit.md`):
```markdown
# Pre-commit Hook

## Actions
1. Run linter
2. Execute tests
3. Check types
4. Format code

## Script
```bash
npm run lint:fix && \
npm run test:changed && \
npm run type-check && \
npm run format
```

## Failure Behavior
- Block commit on test failure
- Auto-fix linting issues
- Report type errors
```

### 3. Built-in Workflow Patterns

**Explore-Plan-Code-Commit Cycle**:
```markdown
# EPCC Workflow

## 1. Explore Phase
- Understand requirements
- Research existing code
- Identify dependencies

## 2. Plan Phase
- Design solution architecture
- Define implementation steps
- Estimate complexity

## 3. Code Phase
- Implement solution
- Write tests
- Add documentation

## 4. Commit Phase
- Review changes
- Write commit message
- Push to repository
```

**Test-Driven Development**:
```markdown
# TDD Workflow

## Red Phase
- Write failing test
- Define expected behavior

## Green Phase
- Implement minimal code
- Make test pass

## Refactor Phase
- Improve code quality
- Maintain test passing
```

### 4. Headless Mode

**CLI Integration**: 
```bash
# Run Claude Code in headless mode
claude -p "Implement user authentication"
```

## MCP (Model Context Protocol) Servers

### Configuration Format

**File Format**: JSON

**Configuration Locations**:
1. **Project Level**: `.mcp.json` in project root
2. **User Level**: `~/.claude.json`
3. **Scope-based**: Via CLI with scope flags

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
      "command": "node",
      "args": ["/usr/local/bin/postgres-mcp"],
      "env": {
        "DATABASE_URL": "${DATABASE_URL}",
        "SSL_MODE": "require"
      },
      "transport": "stdio",
      "permissions": {
        "read": true,
        "write": true,
        "admin": false
      }
    },
    "remote-api": {
      "url": "https://api.example.com/mcp",
      "transport": "http",
      "auth": {
        "type": "oauth",
        "client_id": "${CLIENT_ID}",
        "client_secret": "${CLIENT_SECRET}"
      }
    }
  }
}
```

### CLI Configuration Management

```bash
# Add MCP server
claude mcp add github --scope project

# List configured servers
claude mcp list

# Remove server
claude mcp remove github

# Test server connection
claude mcp test postgres

# Update server configuration
claude mcp update github --env GITHUB_TOKEN=new_token
```

### Supported Transport Protocols

1. **stdio** (Standard Input/Output)
   - Local server processes
   - Most common implementation

2. **SSE** (Server-Sent Events)
   - Streaming responses
   - Real-time updates

3. **HTTP/HTTPS**
   - Remote servers
   - RESTful APIs
   - OAuth authentication support

### OAuth Configuration for Remote Servers

```json
{
  "mcpServers": {
    "corporate-api": {
      "url": "https://mcp.company.com",
      "transport": "http",
      "auth": {
        "type": "oauth",
        "authorization_url": "https://auth.company.com/oauth/authorize",
        "token_url": "https://auth.company.com/oauth/token",
        "client_id": "${OAUTH_CLIENT_ID}",
        "client_secret": "${OAUTH_CLIENT_SECRET}",
        "scope": "read write admin",
        "redirect_uri": "http://localhost:8080/callback"
      },
      "rate_limit": {
        "requests_per_minute": 100,
        "retry_after": true
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
├── CLAUDE.md
├── CLAUDE.local.md (optional overrides)
├── .claude/
│   ├── commands/
│   │   ├── deploy.md
│   │   ├── test.md
│   │   └── migrate.md
│   ├── hooks/
│   │   ├── pre-commit.md
│   │   └── on-save.md
│   └── workflows/
│       ├── tdd.md
│       └── feature-branch.md
├── .mcp.json
├── src/
└── package.json
```

### Example CLAUDE.md: Full-Stack Application

```markdown
# E-Commerce Platform

<system-reminder>
## Project Context
Building a modern e-commerce platform with real-time inventory management.

## Technology Stack
- Frontend: Next.js 14 with App Router
- Backend: Node.js with Express
- Database: PostgreSQL with Prisma
- Cache: Redis
- Queue: Bull MQ
- Monitoring: Sentry

## Architecture Principles
1. Microservices for scalability
2. Event-driven communication
3. CQRS for read/write separation
4. Domain-driven design

## Directory Structure
/apps
  /web - Next.js frontend
  /api - Express backend
  /admin - Admin dashboard
/packages
  /ui - Shared UI components
  /database - Prisma schema
  /types - Shared TypeScript types
/services
  /auth - Authentication service
  /inventory - Inventory management
  /orders - Order processing

## Development Commands
```bash
# Development
pnpm dev          # Start all services
pnpm dev:web      # Frontend only
pnpm dev:api      # Backend only

# Testing
pnpm test         # Run all tests
pnpm test:unit    # Unit tests only
pnpm test:e2e     # E2E tests

# Database
pnpm db:migrate   # Run migrations
pnpm db:seed      # Seed database
pnpm db:studio    # Prisma Studio

# Deployment
pnpm build        # Build all apps
pnpm deploy:staging
pnpm deploy:production
```

## Coding Standards
- TypeScript strict mode enabled
- 100% type coverage required
- Minimum 80% test coverage
- No any types allowed
- Use named exports
- Implement error boundaries

## API Conventions
- RESTful endpoints
- JWT authentication
- Rate limiting enabled
- Request validation with Zod
- Consistent error responses

## Security Requirements
- Input sanitization mandatory
- SQL injection prevention
- XSS protection enabled
- CORS properly configured
- Secrets in environment variables

## Performance Targets
- LCP < 2.5s
- FID < 100ms
- CLS < 0.1
- API response < 200ms
- Database queries < 50ms

## Never Modify
- node_modules/
- .next/
- dist/
- .env files
- Migration files after deployment
</system-reminder>
```

### Example Workflow: Feature Development

```markdown
# Feature Development Workflow

## Trigger
/feature [feature-name]

## Steps

### 1. Setup Feature Branch
```bash
git checkout -b feature/${FEATURE_NAME}
git pull origin main
```

### 2. Create Tests First
- Unit tests for business logic
- Integration tests for APIs
- E2E tests for critical paths

### 3. Implement Feature
- Follow TDD cycle
- Commit frequently
- Update documentation

### 4. Code Review Prep
```bash
npm run lint:fix
npm run test
npm run type-check
npm run build
```

### 5. Create Pull Request
- Add description
- Link to issue
- Request reviewers
- Check CI status

### 6. Post-Merge
```bash
git checkout main
git pull origin main
git branch -d feature/${FEATURE_NAME}
```

## Rollback Procedure
```bash
git revert HEAD
git push origin main
npm run deploy:rollback
```
```

## Best Practices

### CLAUDE.md Organization

1. **Hierarchy**: Use clear section headers
2. **Prioritization**: Most important info first
3. **Specificity**: Project-specific only
4. **Maintenance**: Regular updates
5. **Testing**: Validate with team

### Workflow Design

1. **Modularity**: Small, composable workflows
2. **Error Handling**: Include failure paths
3. **Documentation**: Clear step descriptions
4. **Automation**: Minimize manual steps
5. **Idempotency**: Safe to re-run

### MCP Configuration

1. **Security**: Never hardcode secrets
2. **Permissions**: Principle of least privilege
3. **Testing**: Verify before production
4. **Monitoring**: Log MCP interactions
5. **Documentation**: Setup instructions

## Limitations and Considerations

### System Limitations

- **CLAUDE.md**: Context window impact
- **Workflows**: Terminal interaction required
- **MCP**: Manual tool approval default
- **CLI**: Limited GUI options

### Performance Considerations

- Large CLAUDE.md files affect startup
- Complex workflows may timeout
- MCP server initialization overhead
- Network latency for remote servers

### Security Considerations

- CLAUDE.md visible in repository
- Workflow scripts need validation
- MCP credentials management
- OAuth token refresh handling

## Advanced Features

### Conditional Instructions

```markdown
<system-reminder>
## Environment-Specific Settings

<if environment="development">
- Use local database
- Enable debug logging
- Skip authentication
</if>

<if environment="production">
- Use connection pooling
- Enable caching
- Require authentication
- Monitor performance
</if>
</system-reminder>
```

## Team Collaboration

### Shared Standards

1. **CLAUDE.md Template**: Team-wide template
2. **Command Library**: Shared commands repository
3. **Workflow Patterns**: Documented patterns
4. **MCP Registry**: Approved server list

### Best Practices

1. **Version Control**: Track all configuration
2. **Code Review**: Include CLAUDE.md changes
3. **Documentation**: Maintain setup guides
4. **Training**: Regular team sessions
5. **Feedback**: Continuous improvement