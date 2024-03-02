package yfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"text/template"
)

type Opt func(ctx *DefaultProperties) error

type DefaultProperties struct {
	Value *Value
	Env   map[string]string

	reader ValueReader
	loader ValueLoader

	cache map[string]interface{}
	lock  sync.RWMutex
}

func New(opts ...Opt) *DefaultProperties {
	ret := &DefaultProperties{
		Value:  nil,
		reader: NewYamlReader(),
		loader: NewYamlLoader(),
		cache:  map[string]interface{}{},
	}

	for _, opt := range opts {
		err := opt(ret)
		if err != nil {
			logf("opt err! : %s\n", err.Error())
			return nil
		}
	}

	return ret
}

func (ctx *DefaultProperties) SetValueReader(r ValueReader) {
	ctx.reader = r
}

func (ctx *DefaultProperties) SetValueLoader(l ValueLoader) {
	ctx.loader = l
}

func (ctx *DefaultProperties) ReadValue(r io.Reader) error {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	ctx.cache = map[string]interface{}{}
	ctx.Env = GetEnvs()

	if ctx.reader != nil {
		r, err := ctx.ExecTemplate(r)
		if err != nil {
			return err
		}
		v, err := ctx.reader.Read(r)
		if err != nil {
			return err
		}

		ctx.Value = v
	}
	return nil
}

func GetEnvs() map[string]string {
	s := os.Environ()
	ret := map[string]string{}
	for _, env := range s {
		env := strings.TrimSpace(env)
		if env != "" {
			pair := strings.Split(env, "=")
			if len(pair) == 2 {
				ret[pair[0]] = pair[1]
			}
		}
	}

	return ret
}

func (ctx *DefaultProperties) ExecTemplate(r io.Reader) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	// 替换生产环境下config-prod.yml中env的值
	tpl, ok := template.New("").Option("missingkey=error").Funcs(template.FuncMap{
		"env": ctx.getEnvValue,
	}).Parse(buf.String())
	if ok != nil {
		logf("parse error")
		return nil, ok
	}

	buf.Reset()
	err = tpl.Execute(buf, ctx)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (ctx *DefaultProperties) getEnvValue(key string, arg ...reflect.Value) (reflect.Value, error) {
	if len(key) > 5 && key[:5] == ".Env." {
		key = key[5:]
	}
	if v, ok := ctx.Env[key]; ok {
		return reflect.ValueOf(v), nil
	}

	if len(arg) == 0 {
		return reflect.Value{}, errors.New("no value")
	} else {
		return arg[0], nil
	}
}

// A.B.C
func (ctx *DefaultProperties) Get(key string, defaultValue string) string {
	//if key == "" {
	//	return defaultValue
	//}

	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	if v, ok := ctx.cache[key]; ok {
		if ret, ok := v.(string); ok {
			return ret
		}
	}

	tempKey := "{{ ." + key + "}}"
	tpl, ok := template.New("").Option("missingkey=error").Parse(tempKey)
	if ok != nil {
		logf("key: %s not found(parse error)", key)
		return defaultValue
	}
	b := strings.Builder{}
	err := tpl.Execute(&b, ctx.Value)
	if err != nil {
		return defaultValue
	}

	ret := b.String()
	ctx.cache[key] = ret
	return ret
}

// 依赖于ValueReader的序列化和反序列化方式
func (ctx *DefaultProperties) GetValue(key string, result interface{}) error {
	//if key == "" {
	//	return fmt.Errorf("key is empty")
	//}
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	if v, ok := ctx.cache[key]; ok {
		if ret, ok := v.(string); ok {
			err := ctx.loader.Deserialize(ret, result)
			if err != nil {
				return fmt.Errorf("Unmarshal from cache error: %s, data: %s ", err.Error(), ret)
			}
			return nil
		}
	}

	tempKey := "{{ load_value ." + key + "}}"
	tpl, ok := template.New("").Option("missingkey=error").Funcs(template.FuncMap{
		"load_value": ctx.loader.Serialize,
	}).Parse(tempKey)
	if ok != nil {
		return fmt.Errorf("key: %s not found(parse error)", key)
	}
	b := bytes.NewBuffer(nil)
	err := tpl.Execute(b, ctx.Value)
	if err != nil {
		return fmt.Errorf("load from template failed: err: %s data: %s", err.Error(), b.String())
	}

	data := b.String()
	ctx.cache[key] = data
	err = ctx.loader.Deserialize(data, result)
	if err != nil {
		return fmt.Errorf("Unmarshal error: %s, data: %s ", err.Error(), b.String())
	}
	return nil
}

type JsonReader struct{}

func NewJsonReader() *JsonReader {
	return &JsonReader{}
}

type JsonLoader struct{}

func NewJsonLoader() *JsonLoader {
	return &JsonLoader{}
}

func (v *JsonReader) Read(r io.Reader) (*Value, error) {
	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	ret := Value{}
	logf("value: %s\n", buf.String())
	err = json.Unmarshal(buf.Bytes(), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (v *JsonLoader) Serialize(o interface{}) (string, error) {
	b, err := json.Marshal(o)
	return string(b), err
}

func (v *JsonLoader) Deserialize(value string, result interface{}) error {
	//t := reflect.TypeOf(result)
	//if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.String {
	//	v := reflect.ValueOf(result)
	//	v = v.Elem()
	//	v.SetString(value)
	//	return nil
	//}
	return json.Unmarshal([]byte(value), result)
}
