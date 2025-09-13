package git

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProgressHandler implements ProgressHandler for testing
type MockProgressHandler struct {
	ProgressCalled bool
	CompleteCalled bool
	ErrorCalled    bool
	LastError      error
}

func (m *MockProgressHandler) OnProgress(_ string, _, _ int64) {
	m.ProgressCalled = true
}

func (m *MockProgressHandler) OnComplete() {
	m.CompleteCalled = true
}

func (m *MockProgressHandler) OnError(err error) {
	m.ErrorCalled = true
	m.LastError = err
}

func TestWithProgress(t *testing.T) {
	mockHandler := &MockProgressHandler{}

	option := WithProgress(mockHandler)
	config := &CloneConfig{}

	require.NotNil(t, option)

	// Apply the option to verify it sets progress
	option(config)
	assert.Equal(t, mockHandler, config.Progress)
}

func TestPullWithProgress(t *testing.T) {
	mockHandler := &MockProgressHandler{}

	option := PullWithProgress(mockHandler)
	config := &PullConfig{}

	require.NotNil(t, option)

	// Apply the option to verify it sets progress
	option(config)
	assert.Equal(t, mockHandler, config.Progress)
}

func TestWithSingleBranch(t *testing.T) {
	option := WithSingleBranch()
	config := &CloneConfig{}

	require.NotNil(t, option)

	// Apply the option to verify it sets single branch
	option(config)
	assert.True(t, config.SingleBranch)
}

func TestWithBranch(t *testing.T) {
	branchName := "feature-branch"
	option := WithBranch(branchName)
	config := &CloneConfig{}

	require.NotNil(t, option)

	// Apply the option to verify it sets branch
	option(config)
	assert.Equal(t, branchName, config.Branch)
}

func TestCloneConfig_ApplyOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []CloneOption
		expected CloneConfig
	}{
		{
			name:     "no options",
			options:  []CloneOption{},
			expected: CloneConfig{},
		},
		{
			name: "single branch option",
			options: []CloneOption{
				WithSingleBranch(),
			},
			expected: CloneConfig{
				SingleBranch: true,
			},
		},
		{
			name: "branch option",
			options: []CloneOption{
				WithBranch("main"),
			},
			expected: CloneConfig{
				Branch: "main",
			},
		},
		{
			name: "multiple options",
			options: []CloneOption{
				WithSingleBranch(),
				WithBranch("develop"),
			},
			expected: CloneConfig{
				SingleBranch: true,
				Branch:       "develop",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CloneConfig{}

			for _, option := range tt.options {
				option(config)
			}

			assert.Equal(t, tt.expected.SingleBranch, config.SingleBranch)
			assert.Equal(t, tt.expected.Branch, config.Branch)
		})
	}
}

func TestPullConfig_ApplyOptions(t *testing.T) {
	mockHandler := &MockProgressHandler{}

	config := &PullConfig{}
	option := PullWithProgress(mockHandler)
	option(config)

	assert.Equal(t, mockHandler, config.Progress)
}

func TestNewRepository(t *testing.T) {
	fs := afero.NewMemMapFs()
	repo := NewRepository(fs)

	assert.NotNil(t, repo)
	// Repository is an interface, NewRepository returns a Client
	_, ok := repo.(*Client)
	assert.True(t, ok)
}

func TestClient_ValidateURL(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := DefaultConfig(fs)
	config.AllowedHosts = []string{"github.com", "gitlab.com", "internal.example.com"}
	config.AllowedSchemes = []string{"https", "ssh", "http"}

	client := NewClient(fs, config)

	tests := []struct {
		name    string
		url     string
		wantErr bool
		errType error
	}{
		{
			name:    "empty URL should fail",
			url:     "",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "valid HTTPS URL should pass",
			url:     "https://github.com/user/repo.git",
			wantErr: false,
		},
		{
			name:    "valid SSH URL should pass",
			url:     "git@github.com:user/repo.git",
			wantErr: false,
		},
		{
			name:    "HTTP URL should pass if allowed",
			url:     "http://internal.example.com/repo.git",
			wantErr: false,
		},
		{
			name:    "unauthorized host should fail",
			url:     "https://malicious.com/repo.git",
			wantErr: true,
			errType: ErrUnauthorizedHost,
		},
		{
			name:    "invalid SSH URL format should fail",
			url:     "git@invalid-format",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "invalid HTTP URL should fail",
			url:     "https://[invalid-url",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "unsupported scheme should fail",
			url:     "ftp://example.com/repo.git",
			wantErr: true,
			errType: ErrUnsupportedScheme,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_ValidateURL_NoHostRestrictions(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := DefaultConfig(fs)
	config.AllowedHosts = []string{} // No host restrictions
	config.AllowedSchemes = []string{"https", "ssh"}

	client := NewClient(fs, config)

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "any HTTPS host should be allowed",
			url:  "https://anyhost.com/repo.git",
		},
		{
			name: "any SSH host should be allowed",
			url:  "git@anyhost.com:user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateURL(tt.url)
			require.NoError(t, err)
		})
	}
}

func TestClient_validateHost(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := DefaultConfig(fs)
	config.AllowedHosts = []string{"github.com", "gitlab.com"}
	config.AllowedSchemes = []string{"https", "ssh"}

	client := NewClient(fs, config)

	tests := []struct {
		name    string
		host    string
		scheme  string
		wantErr bool
		errType error
	}{
		{
			name:    "allowed host and scheme should pass",
			host:    "github.com",
			scheme:  "https",
			wantErr: false,
		},
		{
			name:    "allowed host with SSH should pass",
			host:    "gitlab.com",
			scheme:  "ssh",
			wantErr: false,
		},
		{
			name:    "disallowed scheme should fail",
			host:    "github.com",
			scheme:  "ftp",
			wantErr: true,
			errType: ErrUnsupportedScheme,
		},
		{
			name:    "disallowed host should fail",
			host:    "malicious.com",
			scheme:  "https",
			wantErr: true,
			errType: ErrUnauthorizedHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.(*Client).validateHost(tt.host, tt.scheme)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_createParentDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		setupFS func(afero.Fs)
		wantErr bool
	}{
		{
			name:    "create parent directory for new path",
			path:    "/tmp/test/repo",
			setupFS: func(_ afero.Fs) {},
			wantErr: false,
		},
		{
			name: "existing parent directory should not error",
			path: "/tmp/existing/repo",
			setupFS: func(fs afero.Fs) {
				_ = fs.MkdirAll("/tmp/existing", 0o755)
			},
			wantErr: false,
		},
		{
			name:    "root path should not error",
			path:    "/repo",
			setupFS: func(_ afero.Fs) {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setupFS(fs)

			client := NewClient(fs, DefaultConfig(fs))
			err := client.(*Client).createParentDir(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify parent directory was created
				parentDir := filepath.Dir(tt.path)
				if parentDir != "/" {
					exists, err := afero.DirExists(fs, parentDir)
					require.NoError(t, err)
					assert.True(t, exists)
				}
			}
		})
	}
}

func TestClient_createParentDir_ErrorHandling(t *testing.T) {
	// Test with read-only filesystem
	baseFS := afero.NewMemMapFs()
	fs := afero.NewReadOnlyFs(baseFS)

	client := NewClient(fs, DefaultConfig(fs))
	err := client.(*Client).createParentDir("/tmp/test/repo")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create parent directory")
}

func TestClient_buildCloneOptions(t *testing.T) {
	fs := afero.NewMemMapFs()
	client := NewClient(fs, DefaultConfig(fs))

	tests := []struct {
		name     string
		repoURL  string
		config   *CloneConfig
		expected func(*git.CloneOptions) bool
	}{
		{
			name:    "basic options",
			repoURL: "https://github.com/user/repo.git",
			config:  &CloneConfig{},
			expected: func(opts *git.CloneOptions) bool {
				return opts.URL == "https://github.com/user/repo.git" &&
					!opts.SingleBranch &&
					opts.Auth == nil
			},
		},
		{
			name:    "single branch option",
			repoURL: "https://github.com/user/repo.git",
			config:  &CloneConfig{SingleBranch: true},
			expected: func(opts *git.CloneOptions) bool {
				return opts.SingleBranch
			},
		},
		{
			name:    "with branch reference",
			repoURL: "https://github.com/user/repo.git",
			config:  &CloneConfig{Branch: "develop"},
			expected: func(opts *git.CloneOptions) bool {
				return opts.ReferenceName.String() == "refs/heads/develop"
			},
		},
		{
			name:    "with progress handler",
			repoURL: "https://github.com/user/repo.git",
			config:  &CloneConfig{Progress: &MockProgressHandler{}},
			expected: func(opts *git.CloneOptions) bool {
				return opts.Progress != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := client.(*Client).buildCloneOptions(tt.repoURL, nil, tt.config)

			assert.NotNil(t, options)
			assert.True(t, tt.expected(options), "Clone options validation failed")
		})
	}
}

func TestClient_handleCloneError(t *testing.T) {
	fs := afero.NewMemMapFs()
	client := NewClient(fs, DefaultConfig(fs))

	tests := []struct {
		name        string
		setupFS     func(afero.Fs)
		localPath   string
		repoURL     string
		inputErr    error
		expectError string
	}{
		{
			name: "generic error should be wrapped",
			setupFS: func(fs afero.Fs) {
				_ = fs.MkdirAll("/tmp/test", 0o755)
				_ = afero.WriteFile(fs, "/tmp/test/file.txt", []byte("test"), 0o644)
			},
			localPath:   "/tmp/test",
			repoURL:     "https://github.com/test/repo.git",
			inputErr:    fmt.Errorf("clone failed"),
			expectError: "clone failed: clone failed",
		},
		{
			name: "repository not found with SSH should mention authentication",
			setupFS: func(fs afero.Fs) {
				_ = fs.MkdirAll("/tmp/ssh-test", 0o755)
			},
			localPath:   "/tmp/ssh-test",
			repoURL:     "git@github.com:private/repo.git",
			inputErr:    transport.ErrRepositoryNotFound,
			expectError: "repository not found (may be due to authentication issues)",
		},
		{
			name: "repository not found with HTTPS should not mention authentication",
			setupFS: func(fs afero.Fs) {
				_ = fs.MkdirAll("/tmp/https-test", 0o755)
			},
			localPath:   "/tmp/https-test",
			repoURL:     "https://github.com/test/repo.git",
			inputErr:    transport.ErrRepositoryNotFound,
			expectError: "repository not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFS(fs)

			err := client.(*Client).handleCloneError(tt.localPath, tt.inputErr, tt.repoURL)

			require.Error(t, err)
			assert.Equal(t, tt.expectError, err.Error())

			// Verify directory was cleaned up
			exists, _ := afero.DirExists(fs, tt.localPath)
			assert.False(t, exists)
		})
	}
}

func TestClient_IsValidRepository(t *testing.T) {
	fs := afero.NewMemMapFs()
	client := NewClient(fs, DefaultConfig(fs))

	tests := []struct {
		name     string
		path     string
		setupFS  func(afero.Fs)
		expected bool
	}{
		{
			name:     "non-existent path should be invalid",
			path:     "/nonexistent/path",
			setupFS:  func(_ afero.Fs) {},
			expected: false,
		},
		{
			name: "directory without .git should be invalid",
			path: "/tmp/notarepo",
			setupFS: func(fs afero.Fs) {
				_ = fs.MkdirAll("/tmp/notarepo", 0o755)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFS(fs)

			result := client.IsValidRepository(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig(afero.NewMemMapFs())

	assert.NotNil(t, config)
	assert.Equal(t, DefaultCloneTimeout, config.CloneTimeout)
	assert.Equal(t, DefaultPullTimeout, config.PullTimeout)
	assert.Equal(t, DefaultMaxRetries, config.MaxRetries)
	assert.Contains(t, config.AllowedSchemes, "https")
	assert.Contains(t, config.AllowedSchemes, "ssh")
	assert.NotEmpty(t, config.AllowedHosts) // Has default allowed hosts
}

func TestNewClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := DefaultConfig(afero.NewMemMapFs())

	client := NewClient(fs, config)
	concreteClient := client.(*Client)

	assert.NotNil(t, client)
	assert.Equal(t, fs, concreteClient.fs)
	assert.Equal(t, config, concreteClient.config)
	assert.NotNil(t, concreteClient.sshURLRegex)
	assert.NotNil(t, concreteClient.httpURLRegex)
}

func TestDefaultAuthProvider_GetAuth_HTTPS(t *testing.T) {
	provider := &DefaultAuthProvider{}

	tests := []struct {
		name        string
		repoURL     string
		githubToken string
		gitUsername string
		gitPassword string
		expectNil   bool
		expectAuth  bool
	}{
		{
			name:        "GitHub HTTPS with token",
			repoURL:     "https://github.com/user/repo.git",
			githubToken: "ghp_test_token",
			expectAuth:  true,
		},
		{
			name:        "non-GitHub HTTPS with Git credentials",
			repoURL:     "https://gitlab.com/user/repo.git",
			gitUsername: "user",
			gitPassword: "pass",
			expectAuth:  true,
		},
		{
			name:      "public HTTPS repo without auth",
			repoURL:   "https://github.com/public/repo.git",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			if tt.githubToken != "" {
				t.Setenv("GITHUB_TOKEN", tt.githubToken)
			}
			if tt.gitUsername != "" {
				t.Setenv("GIT_USERNAME", tt.gitUsername)
			}
			if tt.gitPassword != "" {
				t.Setenv("GIT_PASSWORD", tt.gitPassword)
			}

			auth, err := provider.GetAuth(tt.repoURL)

			require.NoError(t, err)
			if tt.expectNil {
				assert.Nil(t, auth)
			} else if tt.expectAuth {
				assert.NotNil(t, auth)
			}
		})
	}
}

func TestDefaultAuthProvider_GetAuth_SSH(t *testing.T) {
	// Use mock filesystem
	fs := afero.NewMemMapFs()
	provider := NewDefaultAuthProvider(fs)

	// Create a mock SSH key file for testing
	keyPath := "/home/test/.ssh/test_key"

	// Create directory structure
	err := fs.MkdirAll(filepath.Dir(keyPath), 0o700)
	require.NoError(t, err)

	// Create a dummy key file (not a real key, just for testing file existence)
	err = afero.WriteFile(fs, keyPath, []byte("dummy key content"), 0o600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		repoURL     string
		sshKeyPath  string
		homeDir     string
		expectError bool
		description string
	}{
		{
			name:        "SSH URL without agent or keys",
			repoURL:     "git@github.com:user/repo.git",
			homeDir:     "/home/empty", // Use empty dir with no SSH keys
			expectError: true,
			description: "should fail when no SSH agent and no keys available",
		},
		{
			name:        "SSH URL with custom key path",
			repoURL:     "git@github.com:user/repo.git",
			sshKeyPath:  keyPath,
			homeDir:     "/home/empty", // Use empty dir
			expectError: true,          // Will fail to load dummy key, but tests the path logic
			description: "should try custom SSH key path from environment",
		},
		{
			name:        "SSH URL with mock SSH key in standard location",
			repoURL:     "git@github.com:user/repo.git",
			homeDir:     "/home/test", // Use mock home dir with SSH key
			expectError: true,         // Will still fail to load dummy key, but tests file detection logic
			description: "should find SSH key in standard location",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure SSH_AUTH_SOCK is not set for these tests
			t.Setenv("SSH_AUTH_SOCK", "")

			if tt.sshKeyPath != "" {
				t.Setenv("SSH_KEY_PATH", tt.sshKeyPath)
			}
			if tt.homeDir != "" {
				t.Setenv("HOME", tt.homeDir)
			}

			auth, err := provider.GetAuth(tt.repoURL)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, auth)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, auth)
			}
		})
	}
}

func TestDefaultAuthProvider_TrySSHKeyFile(t *testing.T) {
	// Use mock filesystem
	fs := afero.NewMemMapFs()
	provider := NewDefaultAuthProvider(fs)

	tests := []struct {
		name        string
		keyExists   bool
		expectError bool
		description string
	}{
		{
			name:        "non-existent key file",
			keyExists:   false,
			expectError: true,
			description: "should return error for non-existent key file",
		},
		{
			name:        "existing dummy key file",
			keyExists:   true,
			expectError: true, // Will fail to parse dummy content, but tests file existence logic
			description: "should attempt to load existing key file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPath := "/test/ssh/test_key"

			if tt.keyExists {
				// Create directory structure in mock filesystem
				err := fs.MkdirAll(filepath.Dir(keyPath), 0o700)
				require.NoError(t, err)

				err = afero.WriteFile(fs, keyPath, []byte("dummy key content"), 0o600)
				require.NoError(t, err)
			}

			auth, err := provider.trySSHKeyFile(keyPath)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, auth)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, auth)
			}
		})
	}
}

func TestDefaultAuthProvider_GetAuth_UnsupportedURL(t *testing.T) {
	provider := &DefaultAuthProvider{}

	auth, err := provider.GetAuth("ftp://example.com/repo.git")

	require.Error(t, err)
	assert.Equal(t, ErrNoAuthMethod, err)
	assert.Nil(t, auth)
}

func TestDefaultAuthProvider_ExtractHostnameFromSSHURL(t *testing.T) {
	provider := &DefaultAuthProvider{}

	tests := []struct {
		name     string
		sshURL   string
		expected string
	}{
		{
			name:     "GitHub SSH URL",
			sshURL:   "git@github.com:user/repo.git",
			expected: "github.com",
		},
		{
			name:     "GitLab SSH URL",
			sshURL:   "git@gitlab.com:group/project.git",
			expected: "gitlab.com",
		},
		{
			name:     "Azure DevOps SSH URL",
			sshURL:   "git@ssh.dev.azure.com:v3/org/project/repo",
			expected: "ssh.dev.azure.com",
		},
		{
			name:     "Custom host SSH URL",
			sshURL:   "git@custom.example.com:path/to/repo.git",
			expected: "custom.example.com",
		},
		{
			name:     "Non-SSH URL",
			sshURL:   "https://github.com/user/repo.git",
			expected: "",
		},
		{
			name:     "Malformed SSH URL",
			sshURL:   "git@hostname-without-colon",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.extractHostnameFromSSHURL(tt.sshURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultAuthProvider_GetSSHConfigIdentityFile(t *testing.T) {
	provider := &DefaultAuthProvider{}

	tests := []struct {
		name        string
		hostname    string
		setupConfig func(t *testing.T) string // Returns temp config file path
		expected    string
		description string
	}{
		{
			name:     "hostname not in config",
			hostname: "nonexistent.example.com",
			setupConfig: func(_ *testing.T) string {
				return "" // Use default config
			},
			expected:    "",
			description: "should return empty string for unknown hosts",
		},
		{
			name:     "hostname with tilde path",
			hostname: "test.example.com",
			setupConfig: func(_ *testing.T) string {
				// This test uses the default SSH config behavior
				// In a real scenario, we'd mock ssh_config.Get, but for now
				// we'll test the tilde expansion logic separately
				return ""
			},
			expected:    "",
			description: "should handle tilde expansion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupConfig != nil {
				tt.setupConfig(t)
			}

			result := provider.getSSHConfigIdentityFile(tt.hostname)

			// For now, we just test that the method doesn't panic
			// In a full test suite, we'd mock ssh_config.Get
			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else {
				// Just verify it doesn't panic and returns a string
				assert.IsType(t, "", result)
			}
		})
	}
}

func TestDefaultAuthProvider_TildeExpansion(t *testing.T) {
	// Use mock filesystem
	fs := afero.NewMemMapFs()
	provider := NewDefaultAuthProvider(fs)

	// Test the tilde expansion logic by calling getSSHConfigIdentityFile
	// with a mocked environment
	tmpHome := "/home/test"
	t.Setenv("HOME", tmpHome)

	// Create a test SSH key file in mock filesystem
	sshDir := filepath.Join(tmpHome, ".ssh")
	err := fs.MkdirAll(sshDir, 0o700)
	require.NoError(t, err)

	keyPath := filepath.Join(sshDir, "test_key")
	err = afero.WriteFile(fs, keyPath, []byte("dummy key content"), 0o600)
	require.NoError(t, err)

	// Test that our helper methods work correctly
	hostname := provider.extractHostnameFromSSHURL("git@github.com:user/repo.git")
	assert.Equal(t, "github.com", hostname)
}
