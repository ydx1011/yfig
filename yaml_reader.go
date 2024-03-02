package yfig

import (
	"bytes"
	"github.com/ghodss/yaml"
	"io"
)

type YamlReader struct{}

func NewYamlReader() *YamlReader {
	return &YamlReader{}
}

type YamlLoader struct{}

func NewYamlLoader() *YamlLoader {
	return &YamlLoader{}
}

func (v *YamlReader) Read(r io.Reader) (*Value, error) {
	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	ret := Value{}
	logf("value: %s\n", buf.String())
	err = yaml.Unmarshal(buf.Bytes(), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (v *YamlLoader) Serialize(o interface{}) (string, error) {
	b, err := yaml.Marshal(o)
	return string(b), err
}

func (v *YamlLoader) Deserialize(value string, result interface{}) error {
	return yaml.Unmarshal([]byte(value), result)
}
