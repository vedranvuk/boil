package exec

// Data is the top level data structure passed to a Template file.
type Data struct {
	// Vars is a collection of system variables always present on template
	// execution, generated from environment.
	Vars map[string]string
	// UserVars is a collection of variables given by the user during execution.
	UserVars map[string]string
}

// ReplaceAll replaces all known variable placeholders in input string with
// actual values and returns it.
func (self *Data) ReplaceAll(in string) (out string) {
	// "ProjectName": filepath.Base(config.ModulePath),
	// "TargetDir":   config.TargetDir,
	// "GoVersion":   version,
	// "ModulePath":  config.ModulePath,

	// 	out = in
	// 	for k, v := range self {
	// 		out = strings.ReplaceAll(out, "$"+k, v)
	// 	}
	return
}
