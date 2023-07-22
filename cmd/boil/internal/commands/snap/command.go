package snap

type Config struct {
	// Prompt to add each file.
	ConfirmFiles bool
	// Force overwriting template if it already exists.
	Force bool
}

func Run(cfg *Config) error { return nil }
