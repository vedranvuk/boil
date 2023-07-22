package boil

// State is the current configuration state.
type State struct {
	// ConfigFile is the absolute path of loaded config file.
	ConfigFile string
	// Repository is the absolute path of loaded repository.
	Repository string
	// Prompt for missing required Options via stdin.
	Prompt bool
	// Verbose specifies wether to enable verbose output.
	Verbose bool

	// Metamap is the Metamap of the loaded repository.
	Metamap Metamap
}

// Update updates state based on current settings
func (self *State) Update() error {
	return nil
}