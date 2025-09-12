// Package git provides secure Git repository operations with comprehensive error handling,
// configurable timeouts, and proper authentication management.
//
// Example usage:
//
//	fs := afero.NewOsFs()
//	config := DefaultConfig()
//	client := NewClient(fs, config)
//
//	ctx := context.Background()
//	err := client.Clone(ctx, "https://github.com/user/repo.git", "/tmp/repo", WithBranch("main"))
//	if err != nil {
//	    log.Fatal(err)
//	}
package git

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/kevinburke/ssh_config"
	"github.com/spf13/afero"
)

// Configuration constants for timeouts and limits
const (
	// DefaultCloneTimeout is the default timeout for clone operations
	DefaultCloneTimeout = 5 * time.Minute
	// DefaultPullTimeout is the default timeout for pull operations
	DefaultPullTimeout = 2 * time.Minute
	// DefaultMaxRetries is the default number of retries for transient failures
	DefaultMaxRetries = 3
)

// Common errors
var (
	ErrInvalidURL        = errors.New("invalid repository URL")
	ErrUnsupportedScheme = errors.New("unsupported URL scheme")
	ErrUnauthorizedHost  = errors.New("unauthorized host")
	ErrAuthFailed        = errors.New("authentication failed")
	ErrRepositoryExists  = errors.New("repository already exists")
	ErrNotARepository    = errors.New("not a git repository")
	ErrNoAuthMethod      = errors.New("no authentication method available")
)

// Repository defines the interface for Git repository operations
// Following Go conventions, the interface is named after what it represents
type Repository interface {
	Clone(ctx context.Context, repoURL, localPath string, opts ...CloneOption) error
	Pull(ctx context.Context, localPath string, opts ...PullOption) error
	GetLatestCommitHash(localPath, branch string) (string, error)
	GetFileCommitInfo(localPath, filePath, branch string) (*CommitInfo, error)
	GetCommitInfoByHash(localPath, commitHash string) (*CommitInfo, error)
	GetFileAtCommit(localPath, filePath, commitHash string) ([]byte, error)
	ValidateURL(repoURL string) error
	IsValidRepository(localPath string) bool
	GetRemoteURL(localPath string) (string, error)
}

// CommitInfo represents git commit information
type CommitInfo struct {
	Hash string
	Date string
}

// Config holds configuration for Git operations
type Config struct {
	CloneTimeout    time.Duration
	PullTimeout     time.Duration
	MaxRetries      int
	AllowedHosts    []string
	AllowedSchemes  []string
	ProgressHandler ProgressHandler
	AuthProvider    AuthProvider
}

// ProgressHandler defines the interface for handling clone/pull progress
type ProgressHandler interface {
	OnProgress(message string, current, total int64)
	OnComplete()
	OnError(err error)
}

// AuthProvider defines the interface for providing authentication
type AuthProvider interface {
	GetAuth(repoURL string) (transport.AuthMethod, error)
}

// Client implements Repository using go-git with enhanced security and performance
type Client struct {
	fs     afero.Fs
	config Config
	// Pre-compiled regex for performance
	sshURLRegex  *regexp.Regexp
	httpURLRegex *regexp.Regexp
}

// DefaultConfig returns a configuration with secure defaults
func DefaultConfig() Config {
	return Config{
		CloneTimeout:   DefaultCloneTimeout,
		PullTimeout:    DefaultPullTimeout,
		MaxRetries:     DefaultMaxRetries,
		AllowedSchemes: []string{"https", "ssh"},
		AllowedHosts: []string{
			"github.com",
			"gitlab.com",
			"bitbucket.org",
		},
		AuthProvider: &DefaultAuthProvider{},
	}
}

// NewClient creates a new Client with the provided configuration
func NewClient(fs afero.Fs, config Config) Repository {
	return &Client{
		fs:     fs,
		config: config,
		// Pre-compile regex patterns for performance
		sshURLRegex:  regexp.MustCompile(`^git@([^:]+):(.+)$`),
		httpURLRegex: regexp.MustCompile(`^https?://([^/]+)/(.+)$`),
	}
}

// CloneOption configures clone operations
type CloneOption func(*CloneConfig)

// PullOption configures pull operations
type PullOption func(*PullConfig)

// CloneConfig holds configuration for clone operations
type CloneConfig struct {
	Branch       string
	SingleBranch bool
	Shallow      bool
	Depth        int
	Timeout      time.Duration
	Progress     ProgressHandler
}

// PullConfig holds configuration for pull operations
type PullConfig struct {
	Branch   string
	Timeout  time.Duration
	Progress ProgressHandler
}

// WithBranch sets the branch for clone/pull operations
func WithBranch(branch string) CloneOption {
	return func(c *CloneConfig) {
		c.Branch = branch
	}
}

// WithSingleBranch enables single branch clone
func WithSingleBranch() CloneOption {
	return func(c *CloneConfig) {
		c.SingleBranch = true
	}
}

// WithShallow enables shallow clone
func WithShallow(depth int) CloneOption {
	return func(c *CloneConfig) {
		c.Shallow = true
		c.Depth = depth
	}
}

// WithProgress sets a progress handler for clone operations
func WithProgress(handler ProgressHandler) CloneOption {
	return func(c *CloneConfig) {
		c.Progress = handler
	}
}

// WithTimeout sets a custom timeout for clone operations
func WithTimeout(timeout time.Duration) CloneOption {
	return func(c *CloneConfig) {
		c.Timeout = timeout
	}
}

// PullWithBranch sets the branch for pull operations
func PullWithBranch(branch string) PullOption {
	return func(c *PullConfig) {
		c.Branch = branch
	}
}

// PullWithProgress sets a progress handler for pull operations
func PullWithProgress(handler ProgressHandler) PullOption {
	return func(c *PullConfig) {
		c.Progress = handler
	}
}

// PullWithTimeout sets a custom timeout for pull operations
func PullWithTimeout(timeout time.Duration) PullOption {
	return func(c *PullConfig) {
		c.Timeout = timeout
	}
}

// DefaultAuthProvider provides secure authentication for Git operations
type DefaultAuthProvider struct{}

// GetAuth returns appropriate authentication for the given repository URL
func (p *DefaultAuthProvider) GetAuth(repoURL string) (transport.AuthMethod, error) {
	// SSH authentication
	if strings.HasPrefix(repoURL, "git@") {
		// Try SSH agent first
		auth, err := ssh.NewSSHAgentAuth("git")
		if err == nil {
			return auth, nil
		}

		// SSH agent failed, try SSH key files as fallback
		log.Debug("SSH agent authentication failed, trying SSH key files", "error", err)

		// Extract hostname from SSH URL for SSH config lookup
		hostname := p.extractHostnameFromSSHURL(repoURL)

		// Try SSH config-specified identity file first
		if hostname != "" {
			if keyPath := p.getSSHConfigIdentityFile(hostname); keyPath != "" {
				if auth, err := p.trySSHKeyFile(keyPath); err == nil {
					log.Debug("Successfully authenticated with SSH config identity file", "hostname", hostname, "path", keyPath)
					return auth, nil
				}
				log.Debug("Failed to use SSH config identity file", "hostname", hostname, "path", keyPath, "error", err)
			}
		}

		// Try custom SSH key path from environment
		if keyPath := os.Getenv("SSH_KEY_PATH"); keyPath != "" {
			if auth, err := p.trySSHKeyFile(keyPath); err == nil {
				log.Debug("Successfully authenticated with custom SSH key", "path", keyPath)
				return auth, nil
			}
			log.Debug("Failed to use custom SSH key", "path", keyPath, "error", err)
		}

		// Try standard SSH key locations
		standardKeyPaths := []string{
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519"),
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"),
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_ecdsa"),
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_dsa"),
		}

		for _, keyPath := range standardKeyPaths {
			if auth, err := p.trySSHKeyFile(keyPath); err == nil {
				log.Debug("Successfully authenticated with SSH key", "path", keyPath)
				return auth, nil
			}
		}

		return nil, fmt.Errorf("%w: SSH agent authentication failed and no usable SSH keys found", ErrAuthFailed)
	}

	// HTTPS authentication with domain restrictions
	if strings.HasPrefix(repoURL, "http") {
		// GitHub token authentication (restricted to github.com only)
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			parsed, err := url.Parse(repoURL)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to parse URL: %w", ErrInvalidURL, err)
			}
			// Strict host checking - prevent subdomain attacks
			if parsed.Host == "github.com" {
				return &http.BasicAuth{
					Username: "token",
					Password: token,
				}, nil
			}
		}

		// Generic Git credentials (use with caution)
		if username := os.Getenv("GIT_USERNAME"); username != "" {
			password := os.Getenv("GIT_PASSWORD")
			return &http.BasicAuth{
				Username: username,
				Password: password,
			}, nil
		}

		// No authentication needed for public repos - nil auth with nil error is valid
		//nolint:nilnil // Returning nil,nil is appropriate when no auth is needed
		return nil, nil
	}

	return nil, ErrNoAuthMethod
}

// trySSHKeyFile attempts to create SSH authentication using a specific key file
func (p *DefaultAuthProvider) trySSHKeyFile(keyPath string) (transport.AuthMethod, error) {
	// Check if key file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("SSH key file not found: %s", keyPath)
	}

	// Try to load the key without a passphrase first
	auth, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
	if err == nil {
		return auth, nil
	}

	// If loading without passphrase failed, check if it's due to encryption
	// For now, we don't support encrypted keys in the fallback mechanism
	// to avoid prompting for passphrases in non-interactive contexts
	log.Debug("Failed to load SSH key (may be encrypted)", "path", keyPath, "error", err)
	return nil, fmt.Errorf("failed to load SSH key from %s: %w", keyPath, err)
}

// extractHostnameFromSSHURL extracts the hostname from an SSH URL like "git@github.com:user/repo.git"
func (p *DefaultAuthProvider) extractHostnameFromSSHURL(sshURL string) string {
	// SSH URLs are in format: git@hostname:path/to/repo.git
	if strings.HasPrefix(sshURL, "git@") {
		// Remove "git@" prefix
		withoutPrefix := strings.TrimPrefix(sshURL, "git@")
		// Split on ":" to separate hostname from path
		if colonIndex := strings.Index(withoutPrefix, ":"); colonIndex != -1 {
			return withoutPrefix[:colonIndex]
		}
	}
	return ""
}

// getSSHConfigIdentityFile looks up the IdentityFile for a hostname in SSH config
func (p *DefaultAuthProvider) getSSHConfigIdentityFile(hostname string) string {
	// Use ssh_config to look up the IdentityFile for this hostname
	identityFile := ssh_config.Get(hostname, "IdentityFile")
	if identityFile == "" {
		return ""
	}

	// Expand tilde to home directory if needed
	if strings.HasPrefix(identityFile, "~/") {
		homeDir := os.Getenv("HOME")
		if homeDir != "" {
			return filepath.Join(homeDir, identityFile[2:])
		}
	}

	return identityFile
}

// NewRepository creates a new Repository instance with default configuration
// Maintained for backward compatibility
func NewRepository(fs afero.Fs) Repository {
	return NewClient(fs, DefaultConfig())
}

// Clone clones a git repository with comprehensive security and error handling
func (c *Client) Clone(
	ctx context.Context,
	repoURL, localPath string,
	opts ...CloneOption,
) error {
	// Build configuration from options
	config := &CloneConfig{
		Branch:       "main",
		SingleBranch: false,
		Shallow:      false,
		Timeout:      c.config.CloneTimeout,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Validate URL with security checks
	if err := c.ValidateURL(repoURL); err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	// Set up timeout context
	ctx, cancel := c.setupTimeout(ctx, config.Timeout)
	defer cancel()

	// Ensure parent directory exists
	if err := c.createParentDir(localPath); err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	// Set up authentication
	auth, err := c.config.AuthProvider.GetAuth(repoURL)
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	// Build clone options
	cloneOptions := c.buildCloneOptions(repoURL, auth, config)

	// Perform the clone with context
	if err := c.performClone(ctx, localPath, cloneOptions); err != nil {
		return c.handleCloneError(localPath, err, repoURL)
	}

	// Handle post-clone branch checkout if needed
	if err := c.handlePostCloneBranch(localPath, config.Branch); err != nil {
		// Log but don't fail - allows for tags and other refs
		log.Debug("Post-clone branch checkout failed", "branch", config.Branch, "error", err)
	}

	return nil
}

// Pull updates an existing repository with proper error handling
func (c *Client) Pull(ctx context.Context, localPath string, opts ...PullOption) error {
	// Build configuration from options
	config := &PullConfig{
		Timeout: c.config.PullTimeout,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Set up timeout context
	ctx, cancel := c.setupTimeout(ctx, config.Timeout)
	defer cancel()

	// Open repository
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return fmt.Errorf("pull failed: %w: %w", ErrNotARepository, err)
	}

	// Get remote URL for authentication
	remoteURL, err := c.GetRemoteURL(localPath)
	if err != nil {
		return fmt.Errorf("pull failed: failed to get remote URL: %w", err)
	}

	// Set up authentication
	auth, err := c.config.AuthProvider.GetAuth(remoteURL)
	if err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	// Get working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("pull failed: failed to get worktree: %w", err)
	}

	// Checkout branch if specified
	if config.Branch != "" {
		if err := c.checkoutBranch(localPath, config.Branch); err != nil {
			return fmt.Errorf("pull failed: failed to checkout branch %s: %w", config.Branch, err)
		}
	}

	// Build pull options
	pullOptions := &git.PullOptions{
		Auth: auth,
	}
	if config.Progress != nil {
		pullOptions.Progress = &progressWriter{handler: config.Progress}
	}

	// Perform pull
	err = worktree.PullContext(ctx, pullOptions)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("pull failed: %w", err)
	}

	return nil
}

// GetLatestCommitHash returns the latest commit hash for the specified branch
func (c *Client) GetLatestCommitHash(localPath, branch string) (string, error) {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := c.resolveReference(repo, branch)
	if err != nil {
		return "", fmt.Errorf("failed to resolve reference: %w", err)
	}

	return ref.Hash().String(), nil
}

// GetFileCommitInfo returns the latest commit info for a specific file
func (c *Client) GetFileCommitInfo(localPath, filePath, branch string) (*CommitInfo, error) {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := c.resolveReference(repo, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve reference: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	// Use log to find the latest commit that touches this file
	iter, err := repo.Log(&git.LogOptions{
		From:     ref.Hash(),
		FileName: &filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}
	defer iter.Close()

	// Get the most recent commit that touched this file
	fileCommit, err := iter.Next()
	if err != nil {
		// If no commits found for this file, use the latest commit
		fileCommit = commit
	}

	return &CommitInfo{
		Hash: fileCommit.Hash.String(), // Full hash (stored in config)
		Date: fileCommit.Author.When.Format("2 Jan 2006"),
	}, nil
}

// GetCommitInfoByHash returns commit info for a specific commit hash
func (c *Client) GetCommitInfoByHash(localPath, commitHash string) (*CommitInfo, error) {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Parse the commit hash - handle both full and short hashes
	hash := plumbing.NewHash(commitHash)
	if len(commitHash) == 7 {
		// If it's a short hash, we need to resolve it to a full hash
		// For now, we'll assume it's a valid short hash and try to find the commit
		iter, err := repo.CommitObjects()
		if err != nil {
			return nil, fmt.Errorf("failed to get commit objects: %w", err)
		}
		defer iter.Close()

		err = iter.ForEach(func(commit *object.Commit) error {
			if strings.HasPrefix(commit.Hash.String(), commitHash) {
				hash = commit.Hash
				return storer.ErrStop // Stop iteration
			}
			return nil
		})
		if err != nil && !errors.Is(err, storer.ErrStop) {
			return nil, fmt.Errorf("failed to find commit with hash %s: %w", commitHash, err)
		}
	}

	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object for hash %s: %w", commitHash, err)
	}

	return &CommitInfo{
		Hash: commit.Hash.String(), // Full hash (stored in config)
		Date: commit.Author.When.Format("2 Jan 2006"),
	}, nil
}

// GetFileAtCommit reads a file's content at a specific commit without modifying the working directory
func (c *Client) GetFileAtCommit(localPath, filePath, commitHash string) ([]byte, error) {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Parse the commit hash - handle both full and short hashes
	hash := plumbing.NewHash(commitHash)
	if len(commitHash) == 7 {
		// If it's a short hash, we need to resolve it to a full hash
		iter, err := repo.CommitObjects()
		if err != nil {
			return nil, fmt.Errorf("failed to get commit objects: %w", err)
		}
		defer iter.Close()

		err = iter.ForEach(func(commit *object.Commit) error {
			if strings.HasPrefix(commit.Hash.String(), commitHash) {
				hash = commit.Hash
				return storer.ErrStop // Stop iteration
			}
			return nil
		})
		if err != nil && !errors.Is(err, storer.ErrStop) {
			return nil, fmt.Errorf("failed to find commit with hash %s: %w", commitHash, err)
		}
	}

	// Get the commit object
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object for hash %s: %w", commitHash, err)
	}

	// Get the file tree for this commit
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for commit %s: %w", commitHash, err)
	}

	// Get the file from the tree
	file, err := tree.File(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s at commit %s: %w", filePath, commitHash, err)
	}

	// Read the file contents
	reader, err := file.Reader()
	if err != nil {
		return nil, fmt.Errorf("failed to get file reader for %s: %w", filePath, err)
	}
	defer func() {
		_ = reader.Close() // Ignore error since content was already read successfully
	}()

	// Read all content
	content := make([]byte, file.Size)
	_, err = reader.Read(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

// ValidateURL validates a git repository URL with comprehensive security checks
func (c *Client) ValidateURL(repoURL string) error {
	if repoURL == "" {
		return fmt.Errorf("%w: empty repository URL", ErrInvalidURL)
	}

	// Handle SSH URLs with regex validation
	if strings.HasPrefix(repoURL, "git@") {
		matches := c.sshURLRegex.FindStringSubmatch(repoURL)
		if len(matches) != 3 {
			return fmt.Errorf("%w: invalid SSH URL format", ErrInvalidURL)
		}
		host := matches[1]
		return c.validateHost(host, "ssh")
	}

	// Handle HTTP/HTTPS URLs with validation
	if strings.HasPrefix(repoURL, "http://") || strings.HasPrefix(repoURL, "https://") {
		parsed, err := url.Parse(repoURL)
		if err != nil {
			return fmt.Errorf("%w: invalid HTTP URL: %w", ErrInvalidURL, err)
		}

		scheme := "https"
		if parsed.Scheme == "http" {
			scheme = "http"
		}

		return c.validateHost(parsed.Host, scheme)
	}

	return fmt.Errorf("%w: supported schemes are %v", ErrUnsupportedScheme, c.config.AllowedSchemes)
}

// IsValidRepository checks if the path contains a valid git repository
func (c *Client) IsValidRepository(localPath string) bool {
	_, err := git.PlainOpen(localPath)
	return err == nil
}

// GetRemoteURL returns the remote URL of a repository
func (c *Client) GetRemoteURL(localPath string) (string, error) {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Try to get the remote with retries for transient issues
	var remote *git.Remote
	for attempts := range 3 {
		remote, err = repo.Remote("origin")
		if err == nil {
			break
		}
		if attempts < 2 {
			// Small delay between retries
			time.Sleep(10 * time.Millisecond)
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to get origin remote after retries: %w", err)
	}

	config := remote.Config()
	if len(config.URLs) == 0 {
		return "", fmt.Errorf("no URLs configured for origin remote")
	}

	return config.URLs[0], nil
}

// Helper methods for GitClient

// setupTimeout sets up a timeout context if one isn't already set
func (c *Client) setupTimeout(
	ctx context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {} // Return no-op cancel if deadline already set
	}
	return context.WithTimeout(ctx, timeout)
}

// createParentDir ensures the parent directory exists
func (c *Client) createParentDir(localPath string) error {
	parentDir := filepath.Dir(localPath)
	if err := c.fs.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
	}
	return nil
}

// validateHost checks if a host is allowed by the configuration
func (c *Client) validateHost(host, scheme string) error {
	// Check scheme allowlist
	schemeAllowed := false
	for _, allowed := range c.config.AllowedSchemes {
		if scheme == allowed {
			schemeAllowed = true
			break
		}
	}
	if !schemeAllowed {
		return fmt.Errorf(
			"%w: scheme %q not in allowlist %v",
			ErrUnsupportedScheme,
			scheme,
			c.config.AllowedSchemes,
		)
	}

	// Check host allowlist if configured
	if len(c.config.AllowedHosts) > 0 {
		hostAllowed := false
		for _, allowed := range c.config.AllowedHosts {
			if host == allowed {
				hostAllowed = true
				break
			}
		}
		if !hostAllowed {
			return fmt.Errorf(
				"%w: host %q not in allowlist %v",
				ErrUnauthorizedHost,
				host,
				c.config.AllowedHosts,
			)
		}
	}

	return nil
}

// buildCloneOptions constructs git.CloneOptions from configuration
func (c *Client) buildCloneOptions(
	repoURL string,
	auth transport.AuthMethod,
	config *CloneConfig,
) *git.CloneOptions {
	options := &git.CloneOptions{
		URL:          repoURL,
		Auth:         auth,
		SingleBranch: config.SingleBranch,
	}

	// Set branch reference if specified
	if config.Branch != "" && config.Branch != "main" && config.Branch != "master" {
		options.ReferenceName = plumbing.ReferenceName("refs/heads/" + config.Branch)
	}

	// Set shallow clone options
	if config.Shallow && config.Depth > 0 {
		options.Depth = config.Depth
	}

	// Set progress handler
	if config.Progress != nil {
		options.Progress = &progressWriter{handler: config.Progress}
	}

	return options
}

// performClone executes the git clone operation
func (c *Client) performClone(
	ctx context.Context,
	localPath string,
	options *git.CloneOptions,
) error {
	_, err := git.PlainCloneContext(ctx, localPath, false, options)
	return err
}

// handleCloneError processes clone errors and cleans up on failure
func (c *Client) handleCloneError(localPath string, err error, repoURL string) error {
	// Clean up on failure
	if cleanupErr := c.fs.RemoveAll(localPath); cleanupErr != nil {
		log.Warn("Failed to cleanup failed clone directory", "path", localPath, "error", cleanupErr)
	}

	// Translate common git errors to user-friendly messages
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return fmt.Errorf("repository is empty")
	}
	if errors.Is(err, transport.ErrRepositoryNotFound) {
		// For SSH URLs, repository not found might be an authentication issue
		if strings.HasPrefix(repoURL, "git@") {
			return fmt.Errorf("repository not found (may be due to authentication issues)")
		}
		return fmt.Errorf("repository not found")
	}
	if errors.Is(err, transport.ErrAuthenticationRequired) {
		return fmt.Errorf("%w: authentication required", ErrAuthFailed)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("clone operation timed out")
	}

	return fmt.Errorf("clone failed: %w", err)
}

// handlePostCloneBranch handles branch checkout after clone if needed
func (c *Client) handlePostCloneBranch(localPath, branch string) error {
	if branch == "" || branch == "main" || branch == "master" {
		return nil // No special handling needed
	}
	return c.checkoutBranch(localPath, branch)
}

// progressWriter wraps a ProgressHandler to implement io.Writer
type progressWriter struct {
	handler ProgressHandler
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	if pw.handler != nil {
		pw.handler.OnProgress(string(p), 0, 0)
	}
	return len(p), nil
}

// resolveReference resolves a branch name to a git reference
func (c *Client) resolveReference(
	repo *git.Repository,
	branch string,
) (*plumbing.Reference, error) {
	if branch != "" {
		// Get specific branch
		ref, err := repo.Reference(plumbing.ReferenceName("refs/heads/"+branch), true)
		if err != nil {
			// Try as a tag
			ref, err = repo.Reference(plumbing.ReferenceName("refs/tags/"+branch), true)
			if err != nil {
				return nil, fmt.Errorf("failed to find branch or tag %s: %w", branch, err)
			}
		}
		return ref, nil
	}

	// Get HEAD
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}
	return ref, nil
}

// checkoutBranch switches to the specified branch
func (c *Client) checkoutBranch(localPath, branch string) error {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Try to checkout as branch first
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branch),
	})
	if err != nil {
		// Try as tag
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/tags/" + branch),
		})
	}

	return err
}
