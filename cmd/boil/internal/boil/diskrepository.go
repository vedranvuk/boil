package boil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// NewDiskRepository returns a new DiskRepository rooted at root.
func NewDiskRepository(root string) *DiskRepository { return &DiskRepository{root} }

// DiskRepository is a repository that works with a local fileystem.
// It is initialized from an absolute filesystem path or a path relative to the
// current working directory.
type DiskRepository struct {
	root string
}

func (self DiskRepository) Location() string { return self.root }

// LoadMetamap implements Repository.LoadMetamap.
func (self DiskRepository) LoadMetamap() (metamap Metamap, err error) {
	var metadata *Metafile
	metamap = make(Metamap)
	if err = filepath.Walk(self.root, func(path string, info fs.FileInfo, err error) error {

		if !info.IsDir() {
			return nil
		}
		if err != nil {
			return fmt.Errorf("walk error: %w", err)
		}

		if metadata, err = readMeta(filepath.Join(path, MetafileName)); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		if metadata == nil {
			return nil
		}
		if metadata.Path, err = filepath.Rel(self.root, path); err != nil {
			return fmt.Errorf("rel failed: %w", err)
		}

		var key string
		if key = strings.TrimPrefix(path, self.root); key != "" {
			key = strings.TrimPrefix(key, string(os.PathSeparator))
		} else {
			key = "."
		}
		metamap[key] = metadata

		if metadata.Groups == nil {
			return nil
		}
		for _, multi := range metadata.Groups {
			metamap[fmt.Sprintf("%s#%s", key, multi.Name)] = metadata
		}

		return nil
	}); err != nil {
		err = fmt.Errorf("load metamap from directory: %w", err)
	}
	return
}

func (self *DiskRepository) HasMeta(path string) (exists bool, err error) {
	return self.Exists(filepath.Join(path, MetafileName))
}

func (self *DiskRepository) OpenMeta(path string) (meta *Metafile, err error) {
	if meta, err = readMeta(filepath.Join(self.root, path, MetafileName)); meta != nil {
		meta.Path = path
	}
	return
}

func (self *DiskRepository) SaveMeta(meta *Metafile) (err error) {
	if err = self.Mkdir(meta.Path); err != nil {
		return
	}

	var file *os.File
	if file, err = os.OpenFile(
		filepath.Join(self.root, meta.Path, MetafileName),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm,
	); err != nil {
		return fmt.Errorf("open metafile: %w", err)
	}

	var data []byte
	if data, err = json.MarshalIndent(meta, "", "\t"); err != nil {
		return fmt.Errorf("marshal metafile: %w", err)
	}
	if _, err = file.Write(data); err != nil {
		return fmt.Errorf("write metafile: %w", err)
	}

	return nil
}

func (self *DiskRepository) Exists(path string) (exists bool, err error) {
	if _, err = os.Stat(filepath.Join(self.root, path)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (self *DiskRepository) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(self.root, name))
}

func (self *DiskRepository) WriteFile(name string, data []byte) error {
	return os.WriteFile(filepath.Join(self.root, name), data, os.ModePerm)
}

func (self *DiskRepository) Mkdir(path string) error {
	return os.MkdirAll(filepath.Join(self.root, path), os.ModePerm)
}

func (self *DiskRepository) Remove(path string) error {
	return os.RemoveAll(filepath.Join(self.root, path))
}

func readMeta(filename string) (meta *Metafile, err error) {
	var data []byte
	if data, err = os.ReadFile(filename); err != nil {
		return nil, fmt.Errorf("openmeta: %w", err)
	}
	meta = new(Metafile)
	if err = json.Unmarshal(data, meta); err != nil {
		return nil, fmt.Errorf("unmarshal metafile: %w", err)
	}
	return
}
