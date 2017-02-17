package main

import (
	"io"
	"io/ioutil"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/pkg/errors"
)

func loadConfiguration(data io.Reader) (*ast.ObjectList, error) {
	configBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, errors.Wrap(err, "could not read bytes for config")
	}
	root, err := hcl.ParseBytes(configBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to parse config")
	}
	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return nil, errors.New("Config file doesn't contain root object")
	}
	return list, nil
}
