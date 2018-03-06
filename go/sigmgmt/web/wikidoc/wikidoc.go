// Copyright 2018 Anapaya Systems

package wikidoc

import (
	"io/ioutil"

	"gopkg.in/russross/blackfriday.v2"
)

type WikiServer struct{}

func (ws *WikiServer) Load(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	wikiContent := blackfriday.Run(b)
	return wikiContent, nil
}
