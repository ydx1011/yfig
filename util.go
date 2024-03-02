package yfig

import (
	"fmt"
	"os"
)

func LoadJsonFile(filename string) (Properties, error) {
	return LoadFile(filename, NewJsonReader(), NewJsonLoader())
}

func LoadYamlFile(filename string) (Properties, error) {
	return LoadFile(filename, NewYamlReader(), NewYamlLoader())
}

func LoadFile(filename string, reader ValueReader, loader ValueLoader) (Properties, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	prop := New()
	prop.SetValueReader(reader)
	prop.SetValueLoader(loader)
	err = prop.ReadValue(f)
	return prop, err
}

var logf = func(format string, o ...interface{}) {
	fmt.Printf(format, o...)
}
