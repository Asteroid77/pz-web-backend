package mods

import (
	"encoding/json"
	"os"
)

type CacheStore interface {
	Load() (map[string]ModInfo, error)
	Save(map[string]ModInfo) error
}

type FileCacheStore struct {
	Path string
}

func (s FileCacheStore) Load() (map[string]ModInfo, error) {
	f, err := os.Open(s.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m map[string]ModInfo
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	if m == nil {
		m = make(map[string]ModInfo)
	}
	return m, nil
}

func (s FileCacheStore) Save(m map[string]ModInfo) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0o644)
}
