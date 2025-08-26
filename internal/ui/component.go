package ui

// Component defines the interface for all UI components.
type Component interface {
	// Render returns the string representation of the component.
	Render() string
}

// Themeable defines the interface for components that support custom themes.
type Themeable interface {
	// WithTheme sets a custom theme for the component.
	WithTheme(theme Theme) Component
}

// Validatable defines the interface for components that can be validated.
type Validatable interface {
	// Validate checks if the component is in a valid state.
	Validate() error
}
