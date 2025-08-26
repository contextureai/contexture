package dependencies

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "with context",
			ctx:  context.Background(),
		},
		{
			name: "with nil context",
			ctx:  nil,
		},
		{
			name: "with canceled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
		},
		{
			name: "with timeout context",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
				cancel() // Cancel immediately for testing
				return ctx
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := New(tt.ctx)

			assert.NotNil(t, deps)
			assert.NotNil(t, deps.FS)
			assert.NotNil(t, deps.Context)

			// Verify it's an OS filesystem
			_, ok := deps.FS.(*afero.OsFs)
			assert.True(t, ok, "Expected OsFs for production")

			// Verify context handling
			if tt.ctx != nil {
				assert.Equal(t, tt.ctx, deps.Context)
			} else {
				assert.NotNil(t, deps.Context)
			}
		})
	}
}

func TestNewForTesting(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "with context",
			ctx:  context.Background(),
		},
		{
			name: "with nil context",
			ctx:  nil,
		},
		{
			name: "with value context",
			ctx:  context.WithValue(context.Background(), contextKey("test"), "value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := NewForTesting(tt.ctx)

			assert.NotNil(t, deps)
			assert.NotNil(t, deps.FS)
			assert.NotNil(t, deps.Context)

			// Verify it's a memory filesystem
			_, ok := deps.FS.(*afero.MemMapFs)
			assert.True(t, ok, "Expected MemMapFs for testing")

			// Verify context handling
			if tt.ctx != nil {
				assert.Equal(t, tt.ctx, deps.Context)
			} else {
				assert.NotNil(t, deps.Context)
			}

			// Test that the filesystem works correctly
			testFile := "/test.txt"
			testContent := []byte("test content")

			err := afero.WriteFile(deps.FS, testFile, testContent, 0o644)
			require.NoError(t, err)

			content, err := afero.ReadFile(deps.FS, testFile)
			require.NoError(t, err)
			assert.Equal(t, testContent, content)

			// Verify isolation - file should not exist on real FS
			_, err = os.Stat(testFile)
			assert.True(t, os.IsNotExist(err))
		})
	}
}

func TestWithContext(t *testing.T) {
	t.Parallel()
	originalCtx := context.Background()
	deps := New(originalCtx)

	newCtx := context.WithValue(context.Background(), contextKey("key"), "value")
	newDeps := deps.WithContext(newCtx)

	// Verify new instance was created
	assert.NotSame(t, deps, newDeps)

	// Verify context was updated
	assert.Equal(t, newCtx, newDeps.Context)
	assert.NotEqual(t, deps.Context, newDeps.Context)

	// Verify filesystem was preserved
	assert.Equal(t, deps.FS, newDeps.FS)

	// Verify original was not modified
	assert.Equal(t, originalCtx, deps.Context)
}

func TestWithFS(t *testing.T) {
	t.Parallel()
	deps := New(context.Background())
	originalFS := deps.FS

	newFS := afero.NewMemMapFs()
	newDeps := deps.WithFS(newFS)

	// Verify new instance was created
	assert.NotSame(t, deps, newDeps)

	// Verify filesystem was updated
	assert.Equal(t, newFS, newDeps.FS)
	assert.NotEqual(t, deps.FS, newDeps.FS)

	// Verify context was preserved
	assert.Equal(t, deps.Context, newDeps.Context)

	// Verify original was not modified
	assert.Equal(t, originalFS, deps.FS)
}

func TestBuilderPattern(t *testing.T) {
	t.Parallel()
	// Test chaining of builder methods
	ctx := context.WithValue(context.Background(), contextKey("test"), "value")
	memFS := afero.NewMemMapFs()

	deps := New(context.Background()).
		WithContext(ctx).
		WithFS(memFS)

	assert.Equal(t, ctx, deps.Context)
	assert.Equal(t, memFS, deps.FS)
}

func TestFilesystemOperations(t *testing.T) {
	t.Parallel()
	t.Run("production filesystem", func(t *testing.T) {
		deps := New(context.Background())

		// Create a temp directory for testing
		tempDir, err := afero.TempDir(deps.FS, "", "deps-test-")
		require.NoError(t, err)
		defer func() {
			_ = deps.FS.RemoveAll(tempDir)
		}()

		// Test file operations
		testFile := tempDir + "/test.txt"
		testContent := []byte("production test")

		err = afero.WriteFile(deps.FS, testFile, testContent, 0o644)
		require.NoError(t, err)

		exists, err := afero.Exists(deps.FS, testFile)
		require.NoError(t, err)
		assert.True(t, exists)

		content, err := afero.ReadFile(deps.FS, testFile)
		require.NoError(t, err)
		assert.Equal(t, testContent, content)
	})

	t.Run("test filesystem", func(t *testing.T) {
		deps := NewForTesting(context.Background())

		// Test isolation
		testFile := "/isolated-test.txt"
		testContent := []byte("isolated test")

		err := afero.WriteFile(deps.FS, testFile, testContent, 0o644)
		require.NoError(t, err)

		// File should exist in memory FS
		exists, err := afero.Exists(deps.FS, testFile)
		require.NoError(t, err)
		assert.True(t, exists)

		// File should not exist on real FS
		_, err = os.Stat(testFile)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestContextCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	deps := New(ctx)

	// Context should not be canceled yet
	select {
	case <-deps.Context.Done():
		t.Fatal("Context should not be canceled yet")
	default:
		// Good
	}

	// Cancel the context
	cancel()

	// Context should be canceled
	select {
	case <-deps.Context.Done():
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be canceled")
	}
}

func TestImmutability(t *testing.T) {
	t.Parallel()
	// Test that modifications through builder methods don't affect original
	deps1 := New(context.Background())
	ctx2 := context.WithValue(context.Background(), contextKey("key"), "value")
	fs2 := afero.NewMemMapFs()

	deps2 := deps1.WithContext(ctx2)
	deps3 := deps2.WithFS(fs2)

	// Each should be a separate instance
	assert.NotSame(t, deps1, deps2)
	assert.NotSame(t, deps2, deps3)
	assert.NotSame(t, deps1, deps3)

	// Original should be unchanged
	assert.NotEqual(t, deps1.Context, ctx2)
	assert.NotEqual(t, deps1.FS, fs2)

	// Intermediate should have partial changes
	assert.Equal(t, deps2.Context, ctx2)
	assert.NotEqual(t, deps2.FS, fs2)

	// Final should have all changes
	assert.Equal(t, deps3.Context, ctx2)
	assert.Equal(t, deps3.FS, fs2)
}

// contextKey is a custom type for context keys
type contextKey string

// Benchmarks

func BenchmarkNew(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	for range b.N {
		_ = New(ctx)
	}
}

func BenchmarkNewForTesting(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	for range b.N {
		_ = NewForTesting(ctx)
	}
}

func BenchmarkWithContext(b *testing.B) {
	deps := New(context.Background())
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")
	b.ResetTimer()

	for range b.N {
		_ = deps.WithContext(ctx)
	}
}

func BenchmarkWithFS(b *testing.B) {
	deps := New(context.Background())
	fs := afero.NewMemMapFs()
	b.ResetTimer()

	for range b.N {
		_ = deps.WithFS(fs)
	}
}

// Example usage

func ExampleNew() {
	// Create dependencies for production use
	deps := New(context.Background())

	// Use the filesystem
	_ = afero.WriteFile(deps.FS, "/tmp/example.txt", []byte("Hello"), 0o644)

	// Use the context
	select {
	case <-deps.Context.Done():
		// Handle cancellation
	default:
		// Continue processing
	}
	// Output:
}

func ExampleDependencies_WithContext() {
	// Create dependencies with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deps := New(context.Background()).WithContext(ctx)

	// Now operations can respect the timeout
	select {
	case <-deps.Context.Done():
		// Timeout reached
	case <-time.After(1 * time.Second):
		// Continue processing
	}
	// Output:
}
