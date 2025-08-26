package cli

// Help templates are immutable constants
const (
	// DefaultAppHelpTemplate is the default template for application help
	DefaultAppHelpTemplate = `{{.Name}} - {{.Usage}}

{{with .Description}}{{.}}

{{end}}{{if .VisibleCommands}}COMMANDS:
{{range .VisibleCommands}}   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{"\n"}}{{end}}{{end}}{{if .VisibleFlagCategories}}
{{range .VisibleFlagCategories}}{{if .Name}}{{.Name}}

{{end}}{{range .Flags}}   {{.}}
{{end}}{{end}}{{else}}{{if .VisibleFlags}}OPTIONS:
{{range .VisibleFlags}}   {{.}}
{{end}}{{end}}{{end}}`

	// DefaultCommandHelpTemplate is the default template for command help
	DefaultCommandHelpTemplate = `{{.HelpName}} - {{.Usage}}

{{with .Description}}{{.}}

{{end}}USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Args}} [arguments...]{{end}}{{end}}{{end}}

{{if .VisibleFlags}}OPTIONS:
{{range .VisibleFlags}}   {{.}}
{{end}}{{end}}`
)

// GetAppHelpTemplate returns the app help template
// This function ensures immutability by always returning the constant
func GetAppHelpTemplate() string {
	return DefaultAppHelpTemplate
}

// GetCommandHelpTemplate returns the command help template
// This function ensures immutability by always returning the constant
func GetCommandHelpTemplate() string {
	return DefaultCommandHelpTemplate
}
