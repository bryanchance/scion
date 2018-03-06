// Copyright 2018 Anapaya Systems

package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/scionproto/scion/go/lib/common"
)

type Global struct {
	// Path to database
	DBPath string
	// Folder in which generated config files are stored
	OutputDir string
	// 0 for basic interface, >0 for advanced interface
	Features FeatureLevel
	// Path to SIG config files on target machines
	SIGCfgPath string
	// WebAssetRoot for static files
	WebAssetRoot string
}

func LoadConfig(name string) (*Global, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, common.NewBasicError("Unable to read config file", err, "name", name)
	}
	var global *Global
	if err := json.Unmarshal(b, &global); err != nil {
		return nil, common.NewBasicError("Unable to parse JSON", err, "name", name)
	}
	if err := checkFolderExists(global.OutputDir); err != nil {
		return nil, err
	}
	return global, nil
}

func checkFolderExists(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return common.NewBasicError("Stat error", err, "path", path)
	}
	if !stat.IsDir() {
		return common.NewBasicError("Path is not a directory", nil, "path", path)
	}
	return nil
}

// FeatureLevel provides access to the feature level via methods. This makes
// feature level tests usable at runtime in templates, as opposed to directly
// comparing to constants which is not supported.
type FeatureLevel struct {
	Level int
}

const (
	LevelBasic int = iota
	LevelAdvanced
)

func (fl *FeatureLevel) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &fl.Level)
}

// StrictBasic returns true if the features level is exactly BasicLevel.
func (fl *FeatureLevel) StrictBasic() bool {
	return fl.Level == LevelBasic
}

// Advanced returns true if the features level is greater or equal than
// AdvancedLevel.
func (fl *FeatureLevel) Advanced() bool {
	return fl.Level >= LevelAdvanced
}
