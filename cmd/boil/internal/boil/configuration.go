// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package boil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"text/tabwriter"

	"github.com/adrg/xdg"
)

const (
	// ConfigDir is default Boil configuration directory name.
	ConfigDir = "boil"
	// ConfigFilename is defualt Boil configuration file name.
	ConfigFilename = "config.json"
	// RepositoryDir is default Boil repository irectory name.
	RepositoryDir = "repository"
)

// DefaultConfigFilename returns the absolute path of default config filename.
func DefaultConfigFilename() string {
	return filepath.Join(DefaultConfigDir(), ConfigFilename)
}

// DefaultConfigDir returns the absolute path of default config directory.
func DefaultConfigDir() string {
	return filepath.Join(xdg.ConfigHome, ConfigDir)
}

// DefaultRepositoryDir returns the absolute path of default repository directory.
func DefaultRepositoryDir() string {
	return filepath.Join(DefaultConfigDir(), RepositoryDir)
}

// DefaultConfig returns a config set to defaults or an error.
func DefaultConfig() (config *Config, err error) {

	var usr *user.User
	if usr, err = user.Current(); err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}
	var name string
	if name = usr.Name; name == "" {
		name = usr.Username
	}

	config = &Config{
		DefaultAuthor: Author{
			Name: name,
		},
		ExternalEditor: Action{
			Program: "code",
			Arguments: []string{
				"-n",
				"$" + VarEditTarget.String(),
			},
			Environment: make(map[string]string),
		},
		RepositoryPath: DefaultRepositoryDir(),
	}
	return
}

// Config represents Boil configuration file.
type Config struct {
	// Author is the default template author info.
	DefaultAuthor Author `json:"defaultAuthor,omitempty"`
	// RepositoryPath is the absolute path to the default repository.
	RepositoryPath string `json:"repositoryPath"`

	// DisableBackup, if true disables output directory backup before
	// Template execution.
	//
	// If backup is disabled, if errors occur during template execution
	// the output directory might contain an incomplete and invalid output.
	DisableBackup bool `json:"disableBackup"`

	// ExternalEditor defines the action to execute for the "edit" command, i.e.
	// an external application to edit the template files and metafile.
	//
	// If no editor is defined Boil opens the Template directory in the default
	// system file explorer.
	ExternalEditor Action `json:"editor,omitempty"`

	// Overrides are the configuration overrides specified on command line.
	// They exist at runtime only and are not serialized with Config.
	Overrides struct {
		// ConfigFile is the absolute path of loaded config file.
		ConfigFile string
		// RepositoryPath is the absolute path of loaded repository.
		RepositoryPath string
		// DisableBackup overrides the Configuration.DisableBackup.
		DisableBackup bool
		// Verbose specifies wether to enable verbose output.
		Verbose bool
	} `json:"-"`

	// Runtime holds the runtime variables.
	// They are set by Command Run functions.
	// They exist at runtime only and are not serialized with Config.
	Runtime struct {
		// LoadedConfigFile is the name of the configuration file last loaded
		// into self using self.LoadFromFile.
		LoadedConfigFile string
	} `json:"-"`
}

// Print prints self to stdout.
func (self *Config) Print() {
	var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
	fmt.Fprintf(wr, "DefaultAuthor.Name\t%s\n", self.DefaultAuthor.Name)
	fmt.Fprintf(wr, "DefaultAuthor.Email\t%s\n", self.DefaultAuthor.Email)
	fmt.Fprintf(wr, "DefaultAuthor.Homepage\t%s\n", self.DefaultAuthor.Homepage)
	fmt.Fprintf(wr, "RepositoryPath\t%s\n", self.GetRepositoryPath())
	fmt.Fprintf(wr, "DisableBackup\t%t\n", self.DisableBackup)
	fmt.Fprintf(wr, "Editor.Program\t%s\n", self.ExternalEditor.Program)
	fmt.Fprintf(wr, "Editor.Arguments\t%v\n", self.ExternalEditor.Arguments)
	wr.Flush()
}

// LoadFromFile loads self from filename or returns an error.
func (self *Config) LoadFromFile(filename string) (err error) {
	var buf []byte
	if buf, err = os.ReadFile(filename); err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	if err = json.Unmarshal(buf, self); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	self.Runtime.LoadedConfigFile = filename
	return
}

// Load loads self from a config file.
// If Self.Overrides.ConfigFile is set, that path is used, otherwise the config
// is loaded from the default config file. If the function fails it returns an
// error.
//
// TODO: Try loading first from program directory on Windows.
func (self *Config) LoadOrCreate() (err error) {
	var fn string
	if fn = DefaultConfigFilename(); self.Overrides.ConfigFile != "" {
		fn = self.Overrides.ConfigFile
	}
	if _, err = os.Stat(fn); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat default config: %w", err)
		}
		return self.SaveToFile(DefaultConfigFilename())
	}
	if err = self.LoadFromFile(fn); err != nil {
		err = fmt.Errorf("load config file '%s': %w", fn, err)
	}
	if _, err = os.Stat(self.RepositoryPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat default config: %w", err)
		}
		return os.MkdirAll(self.RepositoryPath, os.ModePerm)
	}
	return nil
}

// SaveToFile saves self to a file specified by filename or returns an error.
func (self *Config) SaveToFile(filename string) (err error) {
	// Create configuration directory if not exists.
	var dir = filepath.Dir(filename)
	if _, err = os.Stat(dir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat config dir: %w", err)
		}
		if err = os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}
	// Create default repository dir if not exists.
	if _, err = os.Stat(self.RepositoryPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat repository: %w", err)
		}
		if err = os.MkdirAll(DefaultRepositoryDir(), os.ModePerm); err != nil {
			return fmt.Errorf("create default repository dir: %w", err)
		}
	}
	// Marshal and save config.
	var buf []byte
	if buf, err = json.MarshalIndent(self, "", "\t"); err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err = os.WriteFile(filename, buf, os.ModePerm); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

// ShouldBackup returns true if self says that a backups should be performed.
func (self *Config) ShouldBackup() (should bool) {
	if should = !self.Overrides.DisableBackup; !should {
		should = !self.DisableBackup
	}
	return
}

// GetRepositoryPath returns the RepositoryPath considering override values.
func (self *Config) GetRepositoryPath() string {
	if self.Overrides.RepositoryPath != "" {
		return self.Overrides.RepositoryPath
	}
	return self.RepositoryPath
}
