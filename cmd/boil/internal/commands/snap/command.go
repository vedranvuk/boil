package snap

type Config struct {
	// Wizard specifies if a template wizard should be used.
	Wizard bool
	// Force overwriting template if it already exists.
	Overwrite bool
}

func Run(cfg *Config) error { return nil }
