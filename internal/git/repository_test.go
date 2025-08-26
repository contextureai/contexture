package git

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
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
	config := DefaultConfig()
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
	config := DefaultConfig()
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
	config := DefaultConfig()
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

			client := NewClient(fs, DefaultConfig())
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

	client := NewClient(fs, DefaultConfig())
	err := client.(*Client).createParentDir("/tmp/test/repo")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create parent directory")
}

func TestClient_buildCloneOptions(t *testing.T) {
	fs := afero.NewMemMapFs()
	client := NewClient(fs, DefaultConfig())

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
	client := NewClient(fs, DefaultConfig())

	tests := []struct {
		name        string
		setupFS     func(afero.Fs)
		localPath   string
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
			inputErr:    fmt.Errorf("clone failed"),
			expectError: "clone failed: clone failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFS(fs)

			err := client.(*Client).handleCloneError(tt.localPath, tt.inputErr)

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
	client := NewClient(fs, DefaultConfig())

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
	config := DefaultConfig()

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
	config := DefaultConfig()

	client := NewClient(fs, config)
	concreteClient := client.(*Client)

	assert.NotNil(t, client)
	assert.Equal(t, fs, concreteClient.fs)
	assert.Equal(t, config, concreteClient.config)
	assert.NotNil(t, concreteClient.sshURLRegex)
	assert.NotNil(t, concreteClient.httpURLRegex)
}
