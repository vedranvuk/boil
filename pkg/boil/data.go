package boil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vedranvuk/bast/pkg/bast"
)

type Data struct {
	Vars Variables
	Bast *bast.Bast
	Json map[string]any
}

func NewData() *Data {
	return &Data{
		Vars: make(Variables),
		Bast: bast.New(),
		Json: make(map[string]any),
	}
}

// StringVar returns a variable value if it exists and its value is a string.
func (self *Data) StringVar(name string) string {
	if v, exists := self.Vars[name]; exists {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func DataFromInputs(vars Variables, goInput, jsonInput []string) (out *Data, err error) {
	out = new(Data)
	out.Vars = vars
	if out.Bast, err = bast.Load(goInput...); err != nil {
		return nil, fmt.Errorf("load go: %w", err)
	}
	for _, ji := range jsonInput {
		var (
			f = filepath.Base(ji)
			d []byte
			j map[string]any
		)
		if d, err = os.ReadFile(ji); err != nil {
			return nil, fmt.Errorf("load json: %w", err)
		}
		if err = json.Unmarshal(d, &j); err != nil {
			return nil, fmt.Errorf("unmarshal json: %w", err)
		}
		out.Json[f] = j
	}
	return
}
