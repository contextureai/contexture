// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressIndicator(t *testing.T) {
	t.Run("new_progress_indicator", func(t *testing.T) {
		pi := NewProgressIndicator("Loading...")
		assert.NotNil(t, pi)
		assert.Equal(t, "Loading...", pi.message)
		assert.False(t, pi.done)
	})

	t.Run("start_and_finish", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		pi := NewProgressIndicator("Processing...")
		pi.Start()
		pi.Finish("Completed successfully")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "Processing...")
		assert.Contains(t, outputStr, "Completed successfully")
		assert.Contains(t, outputStr, "✓")
		assert.True(t, pi.done)
	})

	t.Run("finish_with_error", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		pi := NewProgressIndicator("Processing...")
		pi.Start()
		pi.FinishWithError("Failed to process")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "Processing...")
		assert.Contains(t, outputStr, "Failed to process")
		assert.Contains(t, outputStr, "✗")
		assert.True(t, pi.done)
	})

	t.Run("update_progress", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		pi := NewProgressIndicator("Loading...")
		pi.Start()
		pi.Update(0.5, "50% complete")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "50% complete")
		assert.Equal(t, "50% complete", pi.message)
	})

	t.Run("update_spinner", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		pi := NewProgressIndicator("Loading...")
		pi.Start()
		pi.UpdateSpinner("Still loading...")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "Still loading...")
		assert.Equal(t, "Still loading...", pi.message)
	})

	t.Run("concurrent_access", func(t *testing.T) {
		pi := NewProgressIndicator("Loading...")
		var wg sync.WaitGroup

		// Test concurrent access to ensure thread safety
		for i := range 10 {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				pi.UpdateSpinner("Message " + string(rune('0'+index)))
			}(i)
		}

		wg.Wait()
		// Should not panic due to concurrent access
		assert.False(t, pi.done)
	})

	t.Run("operations_after_done", func(t *testing.T) {
		pi := NewProgressIndicator("Loading...")
		pi.Finish("Done")

		// These operations should be ignored when done
		pi.Start()
		pi.Update(0.8, "Should be ignored")
		pi.UpdateSpinner("Should be ignored")

		assert.True(t, pi.done)
	})
}

func TestBubblesSpinner(t *testing.T) {
	t.Run("new_bubbles_spinner", func(t *testing.T) {
		spinner := NewBubblesSpinner("Testing...")
		assert.NotNil(t, spinner)
		assert.False(t, spinner.done)
		assert.Equal(t, "Testing...", spinner.message)
	})

	t.Run("view_and_stop", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		spinner := NewBubblesSpinner("Testing...")
		view := spinner.View()
		assert.Contains(t, view, "Testing...")

		spinner.Stop("Test completed")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "Test completed")
		assert.Contains(t, outputStr, "✓")
		assert.True(t, spinner.done)
	})

	t.Run("stop_with_error", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		spinner := NewBubblesSpinner("Processing...")
		view := spinner.View()
		assert.Contains(t, view, "Processing...")

		spinner.StopWithError("Processing failed")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "Processing failed")
		assert.Contains(t, outputStr, "✗")
		assert.True(t, spinner.done)
	})

	t.Run("stop_already_done", func(t *testing.T) {
		spinner := NewBubblesSpinner("Testing...")

		spinner.Stop("First stop")
		// Second stop should be ignored
		spinner.Stop("Second stop")

		assert.True(t, spinner.done)
	})

	t.Run("concurrent_stop", func(t *testing.T) {
		spinner := NewBubblesSpinner("Testing...")

		var wg sync.WaitGroup
		// Test concurrent stop calls
		for range 5 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				spinner.Stop("Concurrent stop")
			}()
		}

		wg.Wait()
		assert.True(t, spinner.done)
	})
}

func TestProgressBar(t *testing.T) {
	t.Run("progress_bar_zero_total", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ProgressBar(5, 0, "Invalid")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)

		// Should produce no output when total is 0
		assert.Empty(t, string(output))
	})

	t.Run("progress_bar_partial", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ProgressBar(5, 10, "Half done")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "50%")
		assert.Contains(t, outputStr, "(5/10)")
		assert.Contains(t, outputStr, "Half done")
		assert.Contains(t, outputStr, "█") // Filled part
		assert.Contains(t, outputStr, "░") // Unfilled part
	})

	t.Run("progress_bar_complete", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ProgressBar(10, 10, "Complete")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "100%")
		assert.Contains(t, outputStr, "(10/10)")
		assert.Contains(t, outputStr, "Complete")
		// Should end with newline when complete
		assert.True(t, strings.HasSuffix(outputStr, "\n"))
	})

	t.Run("progress_bar_over_complete", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ProgressBar(15, 10, "Over")

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		assert.Contains(t, outputStr, "150%")
		assert.Contains(t, outputStr, "(15/10)")
		assert.Contains(t, outputStr, "Over")
		// Should end with newline when current >= total
		assert.True(t, strings.HasSuffix(outputStr, "\n"))
	})
}

func TestWithProgress(t *testing.T) {
	t.Run("successful_function", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		called := false
		err := WithProgress("Testing", func() error {
			called = true
			time.Sleep(50 * time.Millisecond) // Simulate work
			return nil
		})

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		require.NoError(t, err)
		assert.True(t, called)
		assert.Contains(t, outputStr, "Testing")
		assert.Contains(t, outputStr, "Testing completed")
		assert.Contains(t, outputStr, "✓")
	})

	t.Run("failing_function", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		testError := assert.AnError
		err := WithProgress("Failing task", func() error {
			return testError
		})

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		require.Error(t, err)
		assert.Equal(t, testError, err)
		assert.Contains(t, outputStr, "Failing task")
		assert.Contains(t, outputStr, "Failing task failed")
		assert.Contains(t, outputStr, "✗")
	})

	t.Run("nil_function", func(t *testing.T) {
		err := WithProgress("Testing", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "progress function cannot be nil")
	})
}

// Benchmark tests for performance
func BenchmarkProgressIndicator(b *testing.B) {
	// Redirect stdout to discard output during benchmarking
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.Run("create_and_finish", func(b *testing.B) {
		//nolint:intrange // b.N is benchmark-specific, not suitable for range
		for i := 0; i < b.N; i++ {
			pi := NewProgressIndicator("Benchmarking...")
			pi.Start()
			pi.Finish("Done")
		}
	})

	b.Run("update_progress", func(b *testing.B) {
		pi := NewProgressIndicator("Benchmarking...")
		pi.Start()
		b.ResetTimer()

		//nolint:intrange // b.N is benchmark-specific, not suitable for range
		for i := 0; i < b.N; i++ {
			pi.Update(float64(i%100)/100.0, "Updating...")
		}
	})
}

func BenchmarkBubblesSpinner(b *testing.B) {
	// Redirect stdout to discard output during benchmarking
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.Run("view_and_stop", func(b *testing.B) {
		//nolint:intrange // b.N is benchmark-specific, not suitable for range
		for i := 0; i < b.N; i++ {
			spinner := NewBubblesSpinner("Benchmarking...")
			_ = spinner.View() // Test view method
			spinner.Stop("Done")
		}
	})
}
