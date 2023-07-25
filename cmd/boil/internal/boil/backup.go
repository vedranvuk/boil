package boil

// CreateBackup creates a backup of a directory using config to determine
// backup location. Returns the backup id and nil on success or an empty string
// and an error otherwise.
func CreateBackup(config *Configuration, dir string) (string, error) {

	return "", nil
}

func RestoreBackup(id string) error { return nil }
