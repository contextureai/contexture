// Package cache provides a simple repository caching mechanism for Contexture.
package cache

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/git"
	"github.com/spf13/afero"
)

const (
	// DefaultCacheDirName is the default directory name for contexture cache
	DefaultCacheDirName = "contexture"
)

// SimpleCache provides cross-session repository caching with human-readable names
type SimpleCache struct {
	fs         afero.Fs
	repository git.Repository
	baseDir    string
}

// NewSimpleCache creates a new simple cache
func NewSimpleCache(fs afero.Fs, repository git.Repository) *SimpleCache {
	baseDir := filepath.Join(os.TempDir(), DefaultCacheDirName)
	return &SimpleCache{
		fs:         fs,
		repository: repository,
		baseDir:    baseDir,
	}
}

// GetRepository retrieves a repository from the cache or clones it if not present.
// It returns the local path to the cached repository without pulling updates.
// Use GetRepositoryWithUpdate if you need to ensure the latest changes are pulled.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - repoURL: The repository URL (supports SSH and HTTPS)
//   - gitRef: The git reference (branch, tag, or commit)
//
// Returns the local filesystem path to the cached repository.
func (c *SimpleCache) GetRepository(ctx context.Context, repoURL, gitRef string) (string, error) {
	return c.getRepository(ctx, repoURL, gitRef, false)
}

// GetRepositoryWithUpdate retrieves a repository and ensures it has the latest changes.
// If the repository is already cached, it pulls the latest updates before returning.
// If the pull fails, it continues with the cached version (logged as a warning).
// Use GetRepository if you only need to access the cached version without updates.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - repoURL: The repository URL (supports SSH and HTTPS)
//   - gitRef: The git reference (branch, tag, or commit)
//
// Returns the local filesystem path to the updated cached repository.
func (c *SimpleCache) GetRepositoryWithUpdate(
	ctx context.Context,
	repoURL, gitRef string,
) (string, error) {
	return c.getRepository(ctx, repoURL, gitRef, true)
}

// getRepository is the shared implementation for both cache access patterns
func (c *SimpleCache) getRepository(
	ctx context.Context,
	repoURL, gitRef string,
	update bool,
) (string, error) {
	cacheKey := c.generateCacheKey(repoURL, gitRef)
	cachePath := filepath.Join(c.baseDir, cacheKey)

	// Check if repository already cached and valid
	if c.isValidRepository(cachePath) {
		if update {
			log.Debug("Updating cached repository", "path", cachePath)
			if err := c.repository.Pull(ctx, cachePath, git.PullWithBranch(gitRef)); err != nil {
				log.Warn(
					"Failed to pull updates, using cached version",
					"path",
					cachePath,
					"error",
					err,
				)
				// Continue with cached version if pull fails
			}
		} else {
			log.Debug("Using cached repository", "path", cachePath)
		}
		return cachePath, nil
	}

	// Repository not cached, need to clone
	return c.cloneRepository(ctx, repoURL, gitRef, cachePath)
}

// cloneRepository handles the shared clone logic
func (c *SimpleCache) cloneRepository(
	ctx context.Context,
	repoURL, gitRef, cachePath string,
) (string, error) {
	// Ensure base directory exists
	if err := c.fs.MkdirAll(c.baseDir, 0o755); err != nil {
		return "", contextureerrors.Wrap(err, "create cache base directory")
	}

	// Clone repository to cache
	log.Debug("Cloning repository to cache", "url", repoURL, "ref", gitRef, "path", cachePath)
	if err := c.repository.Clone(ctx, repoURL, cachePath, git.WithBranch(gitRef)); err != nil {
		// Clean up failed clone
		_ = c.fs.RemoveAll(cachePath)
		return "", contextureerrors.Wrap(err, "clone repository")
	}

	return cachePath, nil
}

// generateCacheKey creates human-readable cache directory name
func (c *SimpleCache) generateCacheKey(repoURL, gitRef string) string {
	// Handle SSH URLs (git@host:path)
	if strings.HasPrefix(repoURL, "git@") {
		// git@github.com:user/repo.git → github.com_user_repo
		re := regexp.MustCompile(`git@([^:]+):(.+)`)
		matches := re.FindStringSubmatch(repoURL)
		if len(matches) == 3 {
			host := matches[1]
			path := strings.TrimSuffix(matches[2], ".git")
			path = strings.ReplaceAll(path, "/", "_")
			return fmt.Sprintf("%s_%s-%s", host, path, gitRef)
		}
	}

	// Handle HTTPS URLs
	if parsed, err := url.Parse(repoURL); err == nil {
		host := parsed.Host
		path := strings.TrimPrefix(parsed.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		path = strings.ReplaceAll(path, "/", "_")
		return fmt.Sprintf("%s_%s-%s", host, path, gitRef)
	}

	// Fallback: sanitize entire URL
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(repoURL, "_")
	return fmt.Sprintf("%s-%s", sanitized, gitRef)
}

// isValidRepository checks if cached repository is valid
func (c *SimpleCache) isValidRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	exists, _ := afero.DirExists(c.fs, gitDir)
	return exists
}
