package api

// Tool describes a tool.
// It contains its name and version as well as the OS and architecture
// for which it was compiled.
type Tool struct {
	Name    string
	Version string
	OS      string
	Arch    string
}
