package domain

import (
	"path/filepath"
	"sort"
	"strings"
)

// RuleNodeType represents the type of a rule tree node
type RuleNodeType string

const (
	// RuleNodeTypeFolder represents a folder node in the rule tree
	RuleNodeTypeFolder RuleNodeType = "folder"
	// RuleNodeTypeRule represents a rule node in the rule tree
	RuleNodeTypeRule RuleNodeType = "rule"
)

// RuleNode represents a node in the rule tree structure
type RuleNode struct {
	Name     string               `json:"name"`
	Type     RuleNodeType         `json:"type"`
	Path     string               `json:"path"`     // Full path from root
	RuleID   string               `json:"rule_id"`  // Only set for rule nodes
	Children map[string]*RuleNode `json:"children"` // Only set for folder nodes
}

// NewRuleTree creates a new rule tree from a list of rule paths
func NewRuleTree(rulePaths []string) *RuleNode {
	root := &RuleNode{
		Name:     "",
		Type:     RuleNodeTypeFolder,
		Path:     "",
		Children: make(map[string]*RuleNode),
	}

	for _, rulePath := range rulePaths {
		addRuleToTree(root, rulePath)
	}

	return root
}

// addRuleToTree adds a single rule path to the tree
func addRuleToTree(root *RuleNode, rulePath string) {
	// Handle empty or separator-only paths
	if rulePath == "" || rulePath == "/" || strings.Trim(rulePath, "/") == "" {
		return
	}

	// Normalize path separators to forward slashes
	rulePath = strings.ReplaceAll(rulePath, "\\", "/")

	// Split the path into components, removing empty parts
	parts := []string{}
	for _, part := range strings.Split(rulePath, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return
	}

	current := root
	currentPath := ""

	// Navigate/create folders for all parts except the last one
	for _, part := range parts[:len(parts)-1] {
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}

		if current.Children[part] == nil {
			current.Children[part] = &RuleNode{
				Name:     part,
				Type:     RuleNodeTypeFolder,
				Path:     currentPath,
				Children: make(map[string]*RuleNode),
			}
		}
		current = current.Children[part]
	}

	// Add the rule file
	ruleName := parts[len(parts)-1]
	fullPath := rulePath
	if currentPath != "" {
		fullPath = currentPath + "/" + ruleName
	}

	current.Children[ruleName] = &RuleNode{
		Name:   ruleName,
		Type:   RuleNodeTypeRule,
		Path:   fullPath,
		RuleID: rulePath, // Store the original rule path as ID
	}
}

// GetChildren returns sorted children of a node (folders first, then rules)
func (n *RuleNode) GetChildren() []*RuleNode {
	if n.Type != RuleNodeTypeFolder {
		return nil
	}

	var folders []*RuleNode
	var rules []*RuleNode

	for _, child := range n.Children {
		if child.Type == RuleNodeTypeFolder {
			folders = append(folders, child)
		} else {
			rules = append(rules, child)
		}
	}

	// Sort folders and rules separately
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name < folders[j].Name
	})
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Name < rules[j].Name
	})

	// Combine folders first, then rules
	result := make([]*RuleNode, 0, len(folders)+len(rules))
	result = append(result, folders...)
	result = append(result, rules...)

	return result
}

// FindNodeByPath finds a node by its path from the root
func (n *RuleNode) FindNodeByPath(path string) *RuleNode {
	if path == "" || path == "." {
		return n
	}

	// Normalize path separators
	path = strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(path, "/")

	current := n
	for _, part := range parts {
		if part == "" {
			continue
		}
		if current.Children == nil {
			return nil
		}
		child, exists := current.Children[part]
		if !exists {
			return nil
		}
		current = child
	}

	return current
}

// GetParentPath returns the parent path of the current node
func (n *RuleNode) GetParentPath() string {
	if n.Path == "" {
		return ""
	}
	return filepath.Dir(n.Path)
}

// GetAllRules returns all rule nodes in the tree (flattened)
func (n *RuleNode) GetAllRules() []*RuleNode {
	var rules []*RuleNode

	if n.Type == RuleNodeTypeRule {
		rules = append(rules, n)
		return rules
	}

	if n.Children != nil {
		for _, child := range n.Children {
			rules = append(rules, child.GetAllRules()...)
		}
	}

	return rules
}

// GetBreadcrumb returns a breadcrumb path for display
func (n *RuleNode) GetBreadcrumb() []string {
	if n.Path == "" {
		return []string{"/"}
	}

	parts := strings.Split(n.Path, "/")
	breadcrumb := []string{"/"}
	breadcrumb = append(breadcrumb, parts...)
	return breadcrumb
}

// ExtractRulePath extracts the rule path from a contexture rule ID
// Handles formats: [contexture:path/rule], [contexture(source):path/rule], [contexture:path/rule,branch], [contexture:path/rule]{variables}
func ExtractRulePath(ruleID string) string {
	if ruleID == "" {
		return ""
	}

	// Remove contexture wrapper: [contexture:...] or [contexture(source):...]
	pathPart := strings.TrimPrefix(ruleID, "[contexture:")
	if strings.HasPrefix(ruleID, "[contexture(") {
		// Handle format: [contexture(source):path/rule]
		parts := strings.SplitN(pathPart, "):", 2)
		if len(parts) == 2 {
			pathPart = parts[1]
		}
	}
	pathPart = strings.TrimSuffix(pathPart, "]")

	// Remove variables part if present (path/rule,branch]{variables} or path/rule]{variables})
	if bracketIdx := strings.Index(pathPart, "]{"); bracketIdx != -1 {
		pathPart = pathPart[:bracketIdx]
	}

	// Remove branch suffix if present (path/rule,branch)
	if commaIdx := strings.Index(pathPart, ","); commaIdx != -1 {
		pathPart = pathPart[:commaIdx]
	}
	return pathPart
}

// ExtractRuleDisplayPath extracts a display-friendly rule path that includes source for custom rules
// For standard rules: returns just the path (e.g., "languages/go/basics")
// For custom source rules: returns "source/path" (e.g., "git@github.com:user/repo/test/rule")
func ExtractRuleDisplayPath(ruleID string) string {
	if ruleID == "" {
		return ""
	}

	// Handle standard format [contexture:path]
	if strings.HasPrefix(ruleID, "[contexture:") {
		// Standard rule - just extract path
		return ExtractRulePath(ruleID)
	}

	// Handle custom source format [contexture(source):path]
	if strings.HasPrefix(ruleID, "[contexture(") {
		// Find the source part
		sourceStart := len("[contexture(")
		sourceEnd := strings.Index(ruleID[sourceStart:], "):")
		if sourceEnd == -1 {
			// Fallback to standard extraction if malformed
			return ExtractRulePath(ruleID)
		}

		source := ruleID[sourceStart : sourceStart+sourceEnd]

		// Extract the path part (reuse existing logic)
		pathPart := ExtractRulePath(ruleID)

		// Only show source prefix for actual custom Git URLs, not default repo references
		if isCustomGitSource(source) {
			// Format source for display: convert Git URLs to user-friendly format
			displaySource := formatSourceForDisplay(source)

			// Combine source and path
			if pathPart != "" {
				return displaySource + "/" + pathPart
			}
			return displaySource
		}

		// For default repository references (like "github", "local"), just return the path
		return pathPart
	}

	// Fallback for any other format
	return ExtractRulePath(ruleID)
}

// isCustomGitSource checks if a source string is a custom Git URL vs a default repo reference
func isCustomGitSource(source string) bool {
	// Custom Git sources are full URLs or SSH addresses
	return strings.HasPrefix(source, "git@") ||
		strings.HasPrefix(source, "https://") ||
		strings.HasPrefix(source, "http://")
}

// formatSourceForDisplay converts a source URL to a display-friendly format
func formatSourceForDisplay(source string) string {
	// Convert SSH URLs like "git@github.com:user/repo.git" to "git@github.com:user/repo"
	if strings.HasPrefix(source, "git@") && strings.HasSuffix(source, ".git") {
		return strings.TrimSuffix(source, ".git")
	}

	// Convert HTTPS URLs like "https://github.com/user/repo.git" to "github.com/user/repo"
	if source, found := strings.CutPrefix(source, "https://"); found {
		return strings.TrimSuffix(source, ".git")
	}

	// Return as-is for other formats
	return source
}
