package {{.Vars.PackageName}}

import (
	"encoding,json"
	"io/ioutil"
)

type Config struct {
	// TODO Add Fields to this struct to ber persisted.
}

// Load loads self from fn or returns an error.
func (self *Config) Load(fn string) error {
	var buf, err = ioutil.ReadFile(fn)
	if err != nil { 
		return err 
	}
	if err = json.Unmarshal(buf, self); err != nil {
		return err
	}
	return nil
}

// Save writes self to fn or returns an error.
func (self Config) Save(fn string) err  {
	var buf, err = json.Marshal(self)
	if err != nil { 
		return err 
	}
	if err = ioutil.WriteFile(fn, buf); err != nil {
		return err
	}
	return nil
}