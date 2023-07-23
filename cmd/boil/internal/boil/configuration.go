package boil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

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

// Config represents Boil configuration file.
type Config struct {
	// Author is the default template author info.
	DefaultAuthor *Author `json:"defaultAuthor,omitempty"`
	// Repository is the absolute path to the default repository.
	Repository string `json:"repository,omitempty"`

	// Overrides are the configuration overrides specified on command line.
	// They exist at runtime only and are not serialized with Config.
	Overrides struct {
		// ConfigFile is the absolute path of loaded config file.
		ConfigFile string
		// RepositoryRoot is the absolute path of loaded repository.
		Repository string
		// Prompt for missing required Options via stdin.
		Prompt bool
		// Verbose specifies wether to enable verbose output.
		Verbose bool
	} `json:"-"`

	// Runtime holds the runtime variables.
	Runtime struct {
		// LoadedConfigFile is the name of the configuration file last loaded
		// into self using self.LoadFromFile.
		LoadedConfigFile string `json:"-"`

		// LoadedRepository is the absolute path to the loaded repository.
		// Value is empty if no repository was loaded.
		LoadedRepository string `json:"-"`
	} `json:"-"`
}

// DefaultConfig returns a config set to defaults or an error.
func DefaultConfig() (config *Config, err error) {

	var usr *user.User
	if usr, err = user.Current(); err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	return &Config{
		DefaultAuthor: &Author{
			Name: usr.Name,
		},
		Repository: DefaultRepositoryDir(),
	}, nil
}

// LoadFromFile loads self from filename or returns an error.
func (self *Config) LoadFromFile(filename string) (err error) {
	var buf []byte
	if buf, err = ioutil.ReadFile(filename); err != nil {
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
	if _, err = os.Stat(self.Repository); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat repository: %w", err)
		}
		if err = os.MkdirAll(DefaultRepositoryDir(), os.ModePerm); err != nil {
			return fmt.Errorf("create default repository dir: %w", err)
		}
	}
	// Marshal and save config.
	var buf []byte
	if buf, err = json.Marshal(self); err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err = ioutil.WriteFile(filename, buf, os.ModePerm); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}
