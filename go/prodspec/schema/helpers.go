// Copyright 2019 Anapaya Systems

package schema

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/env"
)

// LoadUnvalidated loads the prodspec without checking whether it has been validated.
// This function is used only by the validator itself.
func LoadUnvalidated(filename string) (*Layout, string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, "", common.NewBasicError("Cannot read ProdSpec file", err,
			"filename", filename)
	}
	lt := &Layout{}
	md, err := toml.Decode(string(bytes), &lt)
	if err != nil {
		return nil, "", common.NewBasicError("Cannot read ProdSpec file", err,
			"filename", filename)
	}
	undecoded := md.Undecoded()
	if len(undecoded) > 0 {
		return nil, "", common.NewBasicError("Unrecognized TOML element", nil,
			"element", undecoded[0], "filename", filename)
	}
	lt.stringsToPointers()
	hash := sha256.Sum256(bytes)
	return lt, hex.EncodeToString(hash[:]), nil
}

// Load loads prodspec file and checks whether it has been validated.
// If not so it returns an error.
func Load(filename string) (*Layout, error) {
	lt, hash, err := LoadUnvalidated(filename)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadFile("validated")
	if err != nil {
		return nil, common.NewBasicError("Cannot read 'validated' file", err)
	}
	if string(bytes) != hash {
		return nil, common.NewBasicError("Prodspec file is not validated", nil,
			"expected", hash, "found", string(bytes))
	}
	return lt, nil
}

// Save saves the newly created prodspec model into a file.
func (self *Layout) Save(filename string) error {
	self.Generator = filepath.Base(os.Args[0])
	self.GeneratorVersion = env.StartupVersion
	self.GeneratorBuildChain = env.StartupBuildChain
	self.pointersToStrings()
	f, err := os.Create(filename)
	if err != nil {
		return common.NewBasicError("Cannot create ProdSpec file", err,
			"filename", filename)
	}
	encoder := toml.NewEncoder(bufio.NewWriter(f))
	encoder.Indent = "  "
	err = encoder.Encode(self)
	if err != nil {
		return common.NewBasicError("Cannot encode ProdSpec file", err,
			"filename", filename)
	}
	return nil
}
