// Package project provides project configuration management for the Contexture CLI.
//
// The package follows Repository pattern:
//   - ConfigRepository: Handles file I/O operations
//   - RuleMatcher: Handles rule ID parsing and matching
//   - ProjectManager: High-level configuration management
//
// Basic usage:
//
//	manager := project.NewManager(fs)
//	config, err := manager.LoadConfig(basePath)
//	if err != nil {
//	    return err
//	}
//
// Thread safety: All methods are safe for concurrent use as they operate
// on immutable data or use atomic operations for file I/O.
package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/validation"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// Constants for improved maintainability and performance
const (
	// Default file permissions for configuration files and directories
	configFilePermissions = 0o644
	configDirPermissions  = 0o755

	// Default values for configuration cleanup
	defaultParallelFetches = 5
)

// ConfigRepository defines the interface for configuration persistence operations.
// This interface enables easy testing and different storage backends.
type ConfigRepository interface {
	// Load loads configuration from the specified path
	Load(path string) (*domain.Project, error)

	// Save saves configuration to the specified path atomically
	Save(config *domain.Project, path string) error

	// Exists checks if a configuration file exists at the given path
	Exists(path string) (bool, error)

	// DirExists checks if a directory exists at the given path
	DirExists(path string) (bool, error)

	// GetFilesystem returns the underlying filesystem for advanced operations
	GetFilesystem() afero.Fs
}

// RuleMatcher defines the interface for rule ID parsing and matching operations.
// This interface encapsulates the complex rule matching logic and regex compilation.
type RuleMatcher interface {
	// MatchRule checks if two rule IDs match, handling different formats
	MatchRule(ruleID, targetID string) bool

	// ExtractPath extracts the path component from a rule ID
	ExtractPath(ruleID string) (string, error)

	// IsFullFormat checks if a rule ID is in the full [contexture:path] format
	IsFullFormat(ruleID string) bool
}

// ConfigValidator defines the interface for configuration validation
type ConfigValidator interface {
	// ValidateProject validates a project configuration
	ValidateProject(config *domain.Project) error

	// ValidateRuleRef validates a single rule reference
	ValidateRuleRef(ref domain.RuleRef) error
}

// HomeDirectoryProvider defines the interface for getting the home directory.
// This abstraction allows for easier testing and different implementations.
type HomeDirectoryProvider interface {
	// GetHomeDir returns the user's home directory
	GetHomeDir() (string, error)
}

// Manager provides high-level project configuration management operations.
// It orchestrates the various components to provide a clean API for configuration management.
type Manager struct {
	repo         ConfigRepository
	matcher      RuleMatcher
	validator    ConfigValidator
	homeProvider HomeDirectoryProvider
	cleaner      *ConfigCleaner
}

// ConfigCleaner handles the removal of default values from configurations before saving.
// This reduces file size and improves readability of saved configurations.
type ConfigCleaner struct{}

// DefaultConfigRepository provides file-based configuration persistence using afero.Fs.
type DefaultConfigRepository struct {
	fs afero.Fs
}

// DefaultRuleMatcher provides rule ID parsing and matching with compiled regex for performance.
type DefaultRuleMatcher struct {
	regex *regexp.Regexp
}

// DefaultConfigValidator provides configuration validation.
type DefaultConfigValidator struct {
	v validation.Validator
}

// FailsafeConfigValidator provides a fallback validator that fails gracefully
type FailsafeConfigValidator struct {
	err error
}

// DefaultHomeDirectoryProvider provides home directory resolution.
type DefaultHomeDirectoryProvider struct {
	fs afero.Fs
}

// ConfigError represents a configuration operation error with operation context and file path.
// It implements the error interface and supports error unwrapping for error chain traversal.
type ConfigError struct {
	Operation string
	Path      string
	Err       error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config %s failed for %s: %v", e.Operation, e.Path, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewManager creates a new project configuration manager with all dependencies.
// This is the main entry point for using the project configuration system.
func NewManager(fs afero.Fs) *Manager {
	// Create optimized rule matcher with pre-compiled regex
	matcher := &DefaultRuleMatcher{
		regex: domain.RuleIDParsePatternRegex,
	}

	return &Manager{
		repo:         &DefaultConfigRepository{fs: fs},
		matcher:      matcher,
		validator:    newDefaultConfigValidator(),
		homeProvider: &DefaultHomeDirectoryProvider{fs: fs},
		cleaner:      &ConfigCleaner{},
	}
}

// NewManagerForTesting creates a manager with injectable dependencies for testing.
func NewManagerForTesting(
	repo ConfigRepository,
	matcher RuleMatcher,
	validator ConfigValidator,
	homeProvider HomeDirectoryProvider,
) *Manager {
	return &Manager{
		repo:         repo,
		matcher:      matcher,
		validator:    validator,
		homeProvider: homeProvider,
		cleaner:      &ConfigCleaner{},
	}
}

// LoadConfig loads project configuration from the filesystem with proper error handling.
// It tries multiple locations in order of preference and returns detailed error information.
func (m *Manager) LoadConfig(basePath string) (*domain.ConfigResult, error) {
	if strings.TrimSpace(basePath) == "" {
		return nil, contextureerrors.ValidationErrorf("basePath", "cannot be empty")
	}

	// Try .contexture/ directory first (preferred location)
	contexturePath := domain.GetConfigPath(basePath, domain.ConfigLocationContexture)
	if exists, err := m.repo.Exists(contexturePath); err != nil {
		return nil, &ConfigError{
			Operation: "check existence",
			Path:      contexturePath,
			Err:       err,
		}
	} else if exists {
		config, err := m.repo.Load(contexturePath)
		if err != nil {
			return nil, &ConfigError{
				Operation: "load",
				Path:      contexturePath,
				Err:       err,
			}
		}

		if err := m.validator.ValidateProject(config); err != nil {
			return nil, &ConfigError{
				Operation: "validate",
				Path:      contexturePath,
				Err:       err,
			}
		}

		return &domain.ConfigResult{
			Config:   config,
			Location: domain.ConfigLocationContexture,
			Path:     contexturePath,
		}, nil
	}

	// Try project root as fallback
	rootPath := domain.GetConfigPath(basePath, domain.ConfigLocationRoot)
	if exists, err := m.repo.Exists(rootPath); err != nil {
		return nil, &ConfigError{
			Operation: "check existence",
			Path:      rootPath,
			Err:       err,
		}
	} else if exists {
		config, err := m.repo.Load(rootPath)
		if err != nil {
			return nil, &ConfigError{
				Operation: "load",
				Path:      rootPath,
				Err:       err,
			}
		}

		if err := m.validator.ValidateProject(config); err != nil {
			return nil, &ConfigError{
				Operation: "validate",
				Path:      rootPath,
				Err:       err,
			}
		}

		return &domain.ConfigResult{
			Config:   config,
			Location: domain.ConfigLocationRoot,
			Path:     rootPath,
		}, nil
	}

	return nil, &ConfigError{
		Operation: "locate",
		Path:      basePath,
		Err:       errors.New("no configuration file found"),
	}
}

// SaveConfig saves project configuration with atomic writes and validation.
func (m *Manager) SaveConfig(
	config *domain.Project,
	location domain.ConfigLocation,
	basePath string,
) error {
	if config == nil {
		return contextureerrors.ValidationErrorf("config", "cannot be nil")
	}

	if strings.TrimSpace(basePath) == "" {
		return contextureerrors.ValidationErrorf("basePath", "cannot be empty")
	}

	// Validate configuration before saving
	if err := m.validator.ValidateProject(config); err != nil {
		return &ConfigError{
			Operation: "validate",
			Path:      basePath,
			Err:       err,
		}
	}

	configPath := domain.GetConfigPath(basePath, location)

	// Clean configuration to remove defaults
	cleanConfig := m.cleaner.CleanProject(config)

	if err := m.repo.Save(cleanConfig, configPath); err != nil {
		return &ConfigError{
			Operation: "save",
			Path:      configPath,
			Err:       err,
		}
	}

	return nil
}

// InitConfig creates a new project configuration with validation.
func (m *Manager) InitConfig(
	basePath string,
	formats []domain.FormatType,
	location domain.ConfigLocation,
) (*domain.Project, error) {
	if strings.TrimSpace(basePath) == "" {
		return nil, contextureerrors.ValidationErrorf("basePath", "cannot be empty")
	}

	config := &domain.Project{
		Version: 1,
		Formats: make([]domain.FormatConfig, 0, len(formats)),
		Rules:   make([]domain.RuleRef, 0),
	}

	// Add format configurations with validation
	for _, formatType := range formats {
		formatConfig := domain.FormatConfig{
			Type:    formatType,
			Enabled: true,
		}
		config.Formats = append(config.Formats, formatConfig)
	}

	// Save the configuration
	if err := m.SaveConfig(config, location, basePath); err != nil {
		return nil, contextureerrors.Wrap(err, "save initial config")
	}

	return config, nil
}

// AddRule adds a rule reference to the project configuration with proper validation.
// It handles both new rules and updates to existing rules efficiently.
func (m *Manager) AddRule(config *domain.Project, ruleRef domain.RuleRef) error {
	if config == nil {
		return contextureerrors.ValidationErrorf("config", "cannot be nil")
	}

	if err := m.validator.ValidateRuleRef(ruleRef); err != nil {
		return contextureerrors.Wrap(err, "validate rule reference")
	}

	// Check if rule already exists (O(n) is acceptable for typical rule counts)
	for i, existing := range config.Rules {
		if existing.ID == ruleRef.ID {
			// Update existing rule
			config.Rules[i] = ruleRef
			return nil
		}
	}

	// Add new rule
	config.Rules = append(config.Rules, ruleRef)
	return nil
}

// RemoveRule removes a rule reference with optimized matching logic.
// It accepts both full format [contexture:path] and simple format path.
func (m *Manager) RemoveRule(config *domain.Project, ruleID string) error {
	if config == nil {
		return contextureerrors.ValidationErrorf("config", "cannot be nil")
	}

	if strings.TrimSpace(ruleID) == "" {
		return contextureerrors.ValidationErrorf("ruleID", "cannot be empty")
	}

	for i, rule := range config.Rules {
		if m.matcher.MatchRule(ruleID, rule.ID) {
			// Remove rule by slicing (more efficient than preserving order for most use cases)
			config.Rules = append(config.Rules[:i], config.Rules[i+1:]...)
			return nil
		}
	}

	return &ConfigError{
		Operation: "remove rule",
		Path:      ruleID,
		Err:       errors.New("rule not found"),
	}
}

// HasRule checks if a rule exists in the configuration with optimized matching.
func (m *Manager) HasRule(config *domain.Project, ruleID string) bool {
	if config == nil || strings.TrimSpace(ruleID) == "" {
		return false
	}

	for _, rule := range config.Rules {
		if m.matcher.MatchRule(ruleID, rule.ID) {
			return true
		}
	}
	return false
}

// GetConfigLocation determines the best location for configuration with smart defaults.
func (m *Manager) GetConfigLocation(basePath string, preferContexture bool) domain.ConfigLocation {
	if preferContexture {
		return domain.ConfigLocationContexture
	}

	// Check if .contexture directory already exists
	contextureDir := filepath.Join(basePath, domain.GetContextureDir())
	if exists, _ := m.repo.DirExists(contextureDir); exists {
		return domain.ConfigLocationContexture
	}

	// Default to root
	return domain.ConfigLocationRoot
}

// DiscoverLocalRules discovers all local rules in the project's rules directory
func (m *Manager) DiscoverLocalRules(configResult *domain.ConfigResult) ([]domain.RuleRef, error) {
	if configResult == nil {
		return nil, contextureerrors.ValidationErrorf("configResult", "cannot be nil")
	}

	// Determine rules directory based on config location
	var rulesDir string

	switch configResult.Location {
	case domain.ConfigLocationRoot:
		// If config is in project root, rules directory is "rules/"
		// basePath from configResult.Path points to the project root
		basePath := filepath.Dir(configResult.Path)
		rulesDir = filepath.Join(basePath, domain.LocalRulesDir)
	case domain.ConfigLocationContexture:
		// If config is in .contexture/, rules directory is ".contexture/rules/"
		// basePath from configResult.Path points to the .contexture directory
		// So we just need to add the rules subdirectory
		contextureDir := filepath.Dir(configResult.Path)
		rulesDir = filepath.Join(contextureDir, domain.LocalRulesDir)
	case domain.ConfigLocationGlobal:
		// If config is global (~/.contexture/), rules directory is "~/.contexture/rules/"
		globalDir := filepath.Dir(configResult.Path)
		rulesDir = filepath.Join(globalDir, domain.LocalRulesDir)
	default:
		return nil, contextureerrors.ValidationErrorf("configResult.Location", "unknown location: %s", configResult.Location)
	}

	// Check if rules directory exists
	exists, err := m.repo.DirExists(rulesDir)
	if err != nil {
		return nil, contextureerrors.Wrap(err, "check rules directory")
	}

	if !exists {
		log.Debug("No local rules directory found", "path", rulesDir)
		return nil, nil // No local rules directory, return empty slice
	}

	// Discover all .md files in the rules directory
	var localRules []domain.RuleRef
	err = afero.Walk(m.repo.GetFilesystem(), rulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(info.Name(), domain.MarkdownExt) {
			return nil
		}

		// Get relative path from rules directory
		relPath, err := filepath.Rel(rulesDir, path)
		if err != nil {
			return contextureerrors.Wrap(err, "get relative path")
		}

		// Remove .md extension to get rule ID
		ruleID := strings.TrimSuffix(relPath, domain.MarkdownExt)

		// Create RuleRef for local rule
		localRule := domain.RuleRef{
			ID:     ruleID,
			Source: "local",
		}

		localRules = append(localRules, localRule)
		return nil
	})
	if err != nil {
		return nil, contextureerrors.Wrap(err, "walk rules directory")
	}

	log.Debug("Discovered local rules", "count", len(localRules), "directory", rulesDir)
	return localRules, nil
}

// LoadConfigWithLocalRules loads project configuration and automatically includes local rules
func (m *Manager) LoadConfigWithLocalRules(basePath string) (*domain.ConfigResult, error) {
	// Load the base configuration
	configResult, err := m.LoadConfig(basePath)
	if err != nil {
		return nil, err
	}

	// Discover local rules
	localRules, err := m.DiscoverLocalRules(configResult)
	if err != nil {
		return nil, contextureerrors.Wrap(err, "discover local rules")
	}

	// If we have local rules, merge them with existing rules
	if len(localRules) > 0 {
		// Create a copy of the config to avoid modifying the original
		config := *configResult.Config
		config.Rules = make([]domain.RuleRef, len(configResult.Config.Rules)+len(localRules))

		// Copy existing rules first
		copy(config.Rules, configResult.Config.Rules)

		// Add local rules
		copy(config.Rules[len(configResult.Config.Rules):], localRules)

		// Update the config result
		configResult.Config = &config
		log.Debug("Merged local rules with config", "totalRules", len(config.Rules), "localRules", len(localRules))
	}

	return configResult, nil
}

// Implementation of DefaultConfigRepository

// Load loads project configuration from the specified path
func (r *DefaultConfigRepository) Load(path string) (*domain.Project, error) {
	data, err := afero.ReadFile(r.fs, path)
	if err != nil {
		return nil, contextureerrors.Wrap(err, "read config file")
	}

	var config domain.Project
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, contextureerrors.Wrap(err, "parse config file")
	}

	// Apply default values
	if config.Version == 0 {
		config.Version = 1
	}

	return &config, nil
}

// Save saves project configuration to the specified path
func (r *DefaultConfigRepository) Save(config *domain.Project, path string) error {
	// Ensure directory exists
	if err := r.fs.MkdirAll(filepath.Dir(path), configDirPermissions); err != nil {
		return contextureerrors.Wrap(err, "create config directory")
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return contextureerrors.Wrap(err, "marshal config")
	}

	// Atomic write: write to temp file first, then rename
	tempPath := path + ".tmp"
	if err := afero.WriteFile(r.fs, tempPath, data, configFilePermissions); err != nil {
		return contextureerrors.Wrap(err, "write temp config file")
	}

	if err := r.fs.Rename(tempPath, path); err != nil {
		// Clean up temp file on error
		_ = r.fs.Remove(tempPath)
		return contextureerrors.Wrap(err, "rename temp config file")
	}

	return nil
}

// Exists checks if a file exists at the specified path
func (r *DefaultConfigRepository) Exists(path string) (bool, error) {
	return afero.Exists(r.fs, path)
}

// DirExists checks if a directory exists at the specified path
func (r *DefaultConfigRepository) DirExists(path string) (bool, error) {
	return afero.DirExists(r.fs, path)
}

// GetFilesystem returns the underlying filesystem
func (r *DefaultConfigRepository) GetFilesystem() afero.Fs {
	return r.fs
}

// Implementation of DefaultRuleMatcher

// MatchRule checks if the rule ID matches the target ID
func (m *DefaultRuleMatcher) MatchRule(ruleID, targetID string) bool {
	// Direct match
	if ruleID == targetID {
		return true
	}

	// If input is simple format, try to match against rule path
	if !m.IsFullFormat(ruleID) && m.IsFullFormat(targetID) {
		path, err := m.ExtractPath(targetID)
		if err == nil && path == ruleID {
			return true
		}
	}

	// If input is full format but stored is simple, or vice versa
	if m.IsFullFormat(ruleID) && !m.IsFullFormat(targetID) {
		path, err := m.ExtractPath(ruleID)
		if err == nil && path == targetID {
			return true
		}
	}

	return false
}

// ExtractPath extracts the path from a rule ID
func (m *DefaultRuleMatcher) ExtractPath(ruleID string) (string, error) {
	if !m.IsFullFormat(ruleID) {
		return ruleID, nil
	}

	matches := m.regex.FindStringSubmatch(ruleID)
	if len(matches) < 3 {
		return "", contextureerrors.ValidationErrorf("ruleID", "invalid format: %s", ruleID)
	}

	return matches[2], nil
}

// IsFullFormat checks if the rule ID is in full format
func (m *DefaultRuleMatcher) IsFullFormat(ruleID string) bool {
	return strings.HasPrefix(ruleID, "[contexture")
}

// Implementation of DefaultConfigValidator

// newDefaultConfigValidator creates a new config validator
func newDefaultConfigValidator() *DefaultConfigValidator {
	v, err := validation.NewValidator()
	if err != nil {
		// Log the error and return a validator that always fails gracefully
		log.Error("Failed to create validator for config, using failsafe", "error", err)
		return &DefaultConfigValidator{v: &FailsafeConfigValidator{err: err}}
	}
	return &DefaultConfigValidator{v: v}
}

// ValidateProject validates project configuration
func (v *DefaultConfigValidator) ValidateProject(config *domain.Project) error {
	return v.v.ValidateProject(config)
}

// ValidateRuleRef validates a rule reference
func (v *DefaultConfigValidator) ValidateRuleRef(ref domain.RuleRef) error {
	return v.v.ValidateRuleRef(ref)
}

// Implementation of DefaultHomeDirectoryProvider

// GetHomeDir returns the user's home directory
func (p *DefaultHomeDirectoryProvider) GetHomeDir() (string, error) {
	// For now, we'll need to use os.UserHomeDir but this is abstracted
	// so it can be easily mocked or replaced in tests
	homeDir, err := getUserHomeDir()
	if err != nil {
		return "", contextureerrors.Wrap(err, "get user home directory")
	}
	return homeDir, nil
}

// getUserHomeDir is a wrapper around os.UserHomeDir for easier testing
var getUserHomeDir = os.UserHomeDir

// Implementation of ConfigCleaner

// CleanProject removes default values from project configuration before saving.
func (c *ConfigCleaner) CleanProject(config *domain.Project) *domain.Project {
	if config == nil {
		return config
	}

	// Create a copy to avoid modifying the original
	cleanConfig := &domain.Project{
		Version: config.Version,
		Rules:   make([]domain.RuleRef, 0, len(config.Rules)), // Use 0 length, capacity for filtering
		Formats: make([]domain.FormatConfig, len(config.Formats)),
	}

	// Clean rules - exclude local rules (they should not be saved to config)
	for _, rule := range config.Rules {
		// Skip local rules - they are auto-discovered and should not be persisted
		if rule.Source == "local" {
			continue
		}
		cleanRule := domain.RuleRef{
			ID: rule.ID,
		}

		// Only include source if it's not the default
		if rule.Source != "" && rule.Source != domain.DefaultRepository {
			cleanRule.Source = rule.Source
		}

		// Only include ref if it's not the default
		if rule.Ref != "" && rule.Ref != domain.DefaultBranch {
			cleanRule.Ref = rule.Ref
		}

		// Always include variables if they exist
		if len(rule.Variables) > 0 {
			cleanRule.Variables = rule.Variables
		}

		// Always include commitHash if it exists (needed for update tracking)
		if rule.CommitHash != "" {
			cleanRule.CommitHash = rule.CommitHash
		}

		cleanConfig.Rules = append(cleanConfig.Rules, cleanRule)
	}

	// Clean formats
	for i, format := range config.Formats {
		cleanFormat := domain.FormatConfig{
			Type:    format.Type,
			Enabled: format.Enabled, // Always include enabled for clarity
		}

		// Only include BaseDir if it's not empty
		if format.BaseDir != "" {
			cleanFormat.BaseDir = format.BaseDir
		}

		cleanConfig.Formats[i] = cleanFormat
	}

	// Clean optional fields
	cleanConfig.Providers = c.cleanProviders(config.Providers)
	cleanConfig.Generation = c.cleanGenerationConfig(config.Generation)

	return cleanConfig
}

// cleanProviders cleans provider configurations by removing default values.
func (c *ConfigCleaner) cleanProviders(providers []domain.Provider) []domain.Provider {
	if len(providers) == 0 {
		return nil
	}

	var cleanProviders []domain.Provider
	for _, provider := range providers {
		cleanProvider := domain.Provider{
			Name: provider.Name,
			URL:  provider.URL,
		}

		// Only include defaultBranch if it's not the default
		if provider.DefaultBranch != "" && provider.DefaultBranch != domain.DefaultBranch {
			cleanProvider.DefaultBranch = provider.DefaultBranch
		}

		// Only include auth if it exists
		if provider.Auth != nil {
			cleanProvider.Auth = provider.Auth
		}

		cleanProviders = append(cleanProviders, cleanProvider)
	}

	if len(cleanProviders) > 0 {
		return cleanProviders
	}
	return nil
}

// cleanGenerationConfig cleans generation configuration by removing default values.
func (c *ConfigCleaner) cleanGenerationConfig(
	config *domain.GenerationConfig,
) *domain.GenerationConfig {
	if config == nil {
		return nil
	}

	cleanGen := &domain.GenerationConfig{}
	hasNonDefaults := false

	if config.ParallelFetches > 0 && config.ParallelFetches != defaultParallelFetches {
		cleanGen.ParallelFetches = config.ParallelFetches
		hasNonDefaults = true
	}

	if config.DefaultBranch != "" && config.DefaultBranch != domain.DefaultBranch {
		cleanGen.DefaultBranch = config.DefaultBranch
		hasNonDefaults = true
	}

	if config.CacheEnabled {
		cleanGen.CacheEnabled = config.CacheEnabled
		hasNonDefaults = true
	}

	if config.CacheTTL != "" {
		cleanGen.CacheTTL = config.CacheTTL
		hasNonDefaults = true
	}

	if hasNonDefaults {
		return cleanGen
	}
	return nil
}

// LoadGlobalConfig loads the global configuration from ~/.contexture
func (m *Manager) LoadGlobalConfig() (*domain.ConfigResult, error) {
	globalPath, err := m.getGlobalConfigPath()
	if err != nil {
		return nil, contextureerrors.Wrap(err, "get global config path")
	}

	// Check if global config exists
	exists, err := m.repo.Exists(globalPath)
	if err != nil {
		return nil, &ConfigError{
			Operation: "check existence",
			Path:      globalPath,
			Err:       err,
		}
	}

	if !exists {
		// Global config is optional - return empty result
		return &domain.ConfigResult{
			Config:   nil,
			Location: domain.ConfigLocationGlobal,
			Path:     globalPath,
		}, nil
	}

	// Load and validate
	config, err := m.repo.Load(globalPath)
	if err != nil {
		return nil, &ConfigError{
			Operation: "load",
			Path:      globalPath,
			Err:       err,
		}
	}

	if err := m.validator.ValidateProject(config); err != nil {
		return nil, &ConfigError{
			Operation: "validate",
			Path:      globalPath,
			Err:       err,
		}
	}

	return &domain.ConfigResult{
		Config:   config,
		Location: domain.ConfigLocationGlobal,
		Path:     globalPath,
	}, nil
}

// SaveGlobalConfig saves the global configuration
func (m *Manager) SaveGlobalConfig(config *domain.Project) error {
	if config == nil {
		return contextureerrors.ValidationErrorf("config", "cannot be nil")
	}

	// Validate first
	if err := m.validator.ValidateProject(config); err != nil {
		return &ConfigError{
			Operation: "validate",
			Path:      "global",
			Err:       err,
		}
	}

	// Ensure global directory exists
	globalDir, err := m.getGlobalConfigDir()
	if err != nil {
		return contextureerrors.Wrap(err, "get global config dir")
	}

	if err := m.repo.GetFilesystem().MkdirAll(globalDir, configDirPermissions); err != nil {
		return contextureerrors.Wrap(err, "create global config directory")
	}

	// Get global config path
	globalPath, err := m.getGlobalConfigPath()
	if err != nil {
		return contextureerrors.Wrap(err, "get global config path")
	}

	// Clean configuration before saving
	cleanConfig := m.cleaner.CleanProject(config)

	// Save
	if err := m.repo.Save(cleanConfig, globalPath); err != nil {
		return &ConfigError{
			Operation: "save",
			Path:      globalPath,
			Err:       err,
		}
	}

	return nil
}

// InitializeGlobalConfig creates a default global configuration if it doesn't exist
func (m *Manager) InitializeGlobalConfig() error {
	// Check if it already exists
	globalResult, err := m.LoadGlobalConfig()
	if err != nil {
		return err
	}

	if globalResult != nil && globalResult.Config != nil {
		// Already exists, nothing to do
		return nil
	}

	// Create default global config
	defaultConfig := &domain.Project{
		Version: 1,
		Formats: []domain.FormatConfig{
			{Type: domain.FormatClaude, Enabled: true},
			{Type: domain.FormatCursor, Enabled: true},
			{Type: domain.FormatWindsurf, Enabled: true},
		},
		Rules: []domain.RuleRef{},
	}

	return m.SaveGlobalConfig(defaultConfig)
}

// LoadConfigMerged loads both global and project configs and merges them
func (m *Manager) LoadConfigMerged(basePath string) (*domain.MergedConfig, error) {
	// Load global config (optional)
	globalResult, err := m.LoadGlobalConfig()
	if err != nil {
		return nil, contextureerrors.Wrap(err, "load global config")
	}

	// Load project config (required)
	projectResult, err := m.LoadConfig(basePath)
	if err != nil {
		return nil, contextureerrors.Wrap(err, "load project config")
	}

	// Merge configurations
	merged := m.MergeConfigs(globalResult, projectResult)

	return merged, nil
}

// LoadConfigMergedWithLocalRules loads both global and project configs, merges them, and includes local rules
func (m *Manager) LoadConfigMergedWithLocalRules(basePath string) (*domain.MergedConfig, error) {
	// Load global config (optional)
	globalResult, err := m.LoadGlobalConfig()
	if err != nil {
		return nil, contextureerrors.Wrap(err, "load global config")
	}

	// Load project config with local rules (required)
	projectResult, err := m.LoadConfigWithLocalRules(basePath)
	if err != nil {
		return nil, contextureerrors.Wrap(err, "load project config")
	}

	// Merge configurations
	merged := m.MergeConfigs(globalResult, projectResult)

	return merged, nil
}

// MergeConfigs merges global and project configurations
func (m *Manager) MergeConfigs(global, project *domain.ConfigResult) *domain.MergedConfig {
	result := &domain.MergedConfig{
		Project:     project.Config,
		MergedRules: []domain.RuleWithSource{},
	}

	if global != nil {
		result.GlobalConfig = global.Config
	}

	// If no global config, just use project rules
	if global == nil || global.Config == nil {
		for _, rule := range project.Config.Rules {
			result.MergedRules = append(result.MergedRules, domain.RuleWithSource{
				RuleRef:         rule,
				Source:          domain.RuleSourceProject,
				OverridesGlobal: false,
			})
		}
		return result
	}

	// Build maps for quick lookup (O(n) instead of O(n²))
	projectRules := make(map[string]domain.RuleRef)
	for _, rule := range project.Config.Rules {
		normalizedID := m.normalizeRuleID(rule.ID)
		projectRules[normalizedID] = rule
	}

	globalRuleSet := make(map[string]bool)
	for _, globalRule := range global.Config.Rules {
		normalizedID := m.normalizeRuleID(globalRule.ID)
		globalRuleSet[normalizedID] = true
	}

	// Add global rules first (checking for overrides)
	for _, globalRule := range global.Config.Rules {
		normalizedID := m.normalizeRuleID(globalRule.ID)
		if _, overridden := projectRules[normalizedID]; !overridden {
			// Not overridden - add global rule
			result.MergedRules = append(result.MergedRules, domain.RuleWithSource{
				RuleRef:         globalRule,
				Source:          domain.RuleSourceUser,
				OverridesGlobal: false,
			})
		}
	}

	// Add project rules (some may override global)
	for _, projectRule := range project.Config.Rules {
		normalizedID := m.normalizeRuleID(projectRule.ID)

		// Check if this overrides a global rule (O(1) map lookup instead of O(g) loop)
		overridesGlobal := globalRuleSet[normalizedID]

		result.MergedRules = append(result.MergedRules, domain.RuleWithSource{
			RuleRef:         projectRule,
			Source:          domain.RuleSourceProject,
			OverridesGlobal: overridesGlobal,
		})
	}

	return result
}

// normalizeRuleID extracts and normalizes a rule ID for comparison
// Note: Variables are intentionally ignored - rules with the same path but different
// variables are treated as the same rule for override detection. This means:
//
//	Global: @contexture/go{style: "strict"}
//	Project: @contexture/go{style: "relaxed"}
//
// The project rule will override the global rule entirely.
func (m *Manager) normalizeRuleID(ruleID string) string {
	// Use existing RuleMatcher logic to extract path
	path, err := m.matcher.ExtractPath(ruleID)
	if err != nil {
		// Fallback to the ID itself
		return strings.ToLower(ruleID)
	}
	return strings.ToLower(path)
}

// getGlobalConfigDir returns the global contexture directory using the homeProvider
func (m *Manager) getGlobalConfigDir() (string, error) {
	homeDir, err := m.homeProvider.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".contexture"), nil
}

// getGlobalConfigPath returns the global configuration file path using the homeProvider
func (m *Manager) getGlobalConfigPath() (string, error) {
	dir, err := m.getGlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, domain.GetConfigFileName()), nil
}

// FailsafeConfigValidator methods - all return errors due to initialization failure

// ValidateRule returns an error for FailsafeConfigValidator
func (f *FailsafeConfigValidator) ValidateRule(_ *domain.Rule) *domain.ValidationResult {
	return &domain.ValidationResult{
		Valid:    false,
		Errors:   []error{contextureerrors.Wrap(f.err, "validator initialization")},
		Warnings: make([]domain.ValidationWarning, 0),
	}
}

// ValidateRules returns an error for FailsafeConfigValidator
func (f *FailsafeConfigValidator) ValidateRules(rules []*domain.Rule) *validation.BatchResult {
	return &validation.BatchResult{
		TotalRules:  len(rules),
		ValidRules:  0,
		Results:     []*validation.Result{},
		AllValid:    false,
		HasWarnings: false,
	}
}

// ValidateProject returns an error for FailsafeConfigValidator
func (f *FailsafeConfigValidator) ValidateProject(_ *domain.Project) error {
	return contextureerrors.Wrap(f.err, "validator initialization")
}

// ValidateRuleRef returns an error for FailsafeConfigValidator
func (f *FailsafeConfigValidator) ValidateRuleRef(_ domain.RuleRef) error {
	return contextureerrors.Wrap(f.err, "validator initialization")
}

// ValidateRuleID returns an error for FailsafeConfigValidator
func (f *FailsafeConfigValidator) ValidateRuleID(_ string) error {
	return contextureerrors.Wrap(f.err, "validator initialization")
}

// ValidateGitURL returns an error for FailsafeConfigValidator
func (f *FailsafeConfigValidator) ValidateGitURL(_ string) error {
	return contextureerrors.Wrap(f.err, "validator initialization")
}

// ValidateFormatConfig returns an error for FailsafeConfigValidator
func (f *FailsafeConfigValidator) ValidateFormatConfig(_ *domain.FormatConfig) error {
	return contextureerrors.Wrap(f.err, "validator initialization")
}

// ValidateWithContext returns an error for FailsafeConfigValidator
func (f *FailsafeConfigValidator) ValidateWithContext(
	_ context.Context,
	_ any,
	_ string,
) error {
	return contextureerrors.Wrap(f.err, "validator initialization")
}
