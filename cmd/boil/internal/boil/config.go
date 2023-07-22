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
	// ConfigDir is the default boil configuration dir.
	ConfigDir = "boil"
	// ConfigFilename is boild config filename.
	ConfigFilename = "config.json"
	// RepositoryDir is the name of the default repository.
	RepositoryDir = "repository"
)

// Config is the boil configuration.
type Config struct {
	// Author is the default template author info.
	DefaultAuthor *Author
	// Repository is the absolute path to the default repository.
	Repository string
	// State is the current state of the program.
	State *State
}

// DefaultConfig returns a config set to defaults or an error.
func DefaultConfig() (config *Config, err error) {

	var usr *user.User
	if usr, err = user.Current(); err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	config = &Config{
		DefaultAuthor: &Author{
			Name: usr.Name,
		},
		State: &State{
			Metamap: make(Metamap),
		},
	}
	config.Repository = config.DefaultRepositoryDir()
	config.State.ConfigFile = config.DefaultConfigFilename()
	config.State.Repository = config.DefaultRepositoryDir()

	return config, nil
}

// SetDefaults sets defaults to self.
func (self *Config) SetDefaults() *Config {
	return self
}

// DefaultConfigFilename returns the default config filename.
func (self *Config) DefaultConfigFilename() string {
	return filepath.Join(self.DefaultConfigDir(), ConfigFilename)
}

// DefaultConfigDir returns the default config directory.
func (self *Config) DefaultConfigDir() string {
	return filepath.Join(xdg.ConfigHome, ConfigDir)
}

func (self *Config) DefaultRepositoryDir() string {
	return filepath.Join(self.DefaultConfigDir(), RepositoryDir)
}

// LoadFromFile loads self from filename or returns an error.
func (self *Config) LoadFromFile(filename string) (err error) {
	var buf []byte
	if buf, err = ioutil.ReadFile(filename); err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if err = json.Unmarshal(buf, self); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return
}

// LoadFromDefaultFileOrCreate tries to load the config file from the default
// location. If the file does not exit it is created. If an error occurs it is
// returned.
func (self *Config) LoadFromDefaultFileOrCreate() (err error) {
	if _, err = os.Stat(self.DefaultConfigFilename()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat default config: %w", err)
		}
		err = self.SaveToDefaultLocation()
	}
	return
}

// SaveToDefaultLocation saves self to default config file location.
func (self *Config) SaveToDefaultLocation() (err error) {

	// Save state, nil it during marshaling and defer restoring it.
	var state = self.State
	self.State = nil
	defer func() {
		self.State = state
	}()

	var buf []byte
	if buf, err = json.Marshal(self); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	if err = os.MkdirAll(self.DefaultConfigDir(), os.ModePerm); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	var fn = self.DefaultConfigFilename()
	if err = ioutil.WriteFile(fn, buf, os.ModePerm); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	// Create default repository dir if not exists.
	if _, err = os.Stat(self.DefaultRepositoryDir()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat repository: %w", err)
		}
		if err = os.MkdirAll(self.DefaultRepositoryDir(), os.ModePerm); err != nil {
			return fmt.Errorf("create default repository: %w", err)
		}
	}
	return
}

// InitializeState initializes State variables from state settings.
func (self *Config) InitializeState() (err error) {
	// Load config.
	if err = self.LoadFromFile(self.State.ConfigFile); err != nil {
		err = fmt.Errorf("load config file '%s': %w", self.State.ConfigFile, err)
	}
	// Load repository.
	if self.State.Metamap, err = LoadMetamap(self.State.Repository); err != nil {
		err = fmt.Errorf("load repository metamap: %w", err)
	}
	return
}
