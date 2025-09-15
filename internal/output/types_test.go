package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager_ValidFormats(t *testing.T) {
	tests := []struct {
		name   string
		format Format
	}{
		{"default format", FormatDefault},
		{"json format", FormatJSON},
		{"empty format", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.format)
			require.NoError(t, err)
			assert.NotNil(t, manager)
			assert.NotNil(t, manager.writer)
		})
	}
}

func TestNewManager_InvalidFormat(t *testing.T) {
	manager, err := NewManager("invalid")
	require.Error(t, err)
	assert.Nil(t, manager)

	var unsupportedErr *UnsupportedFormatError
	require.ErrorAs(t, err, &unsupportedErr)
	assert.Equal(t, "invalid", unsupportedErr.Format)
}

func TestUnsupportedFormatError(t *testing.T) {
	err := &UnsupportedFormatError{Format: "yaml"}
	expected := "unsupported output format: yaml (supported formats: default, json)"
	assert.Equal(t, expected, err.Error())
}

func TestListMetadata(t *testing.T) {
	metadata := ListMetadata{
		Pattern:       "testing",
		TotalRules:    5,
		FilteredRules: 2,
	}

	assert.Equal(t, "testing", metadata.Pattern)
	assert.Equal(t, 5, metadata.TotalRules)
	assert.Equal(t, 2, metadata.FilteredRules)
}
