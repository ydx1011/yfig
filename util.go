package yfig

import (
	"errors"
	"fmt"
	"github.com/ydx1011/reflection"
	"os"
	"reflect"
	"strings"
)

const (
	TagPrefixName = "figPx"
	TagName       = "fig"
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

// param: prop 属性
// param: result 填充的struct
// result: result如果不为struct的指针返回错误，填充时异常返回错误
func Fill(prop Properties, result interface{}) error {
	return FillEx(prop, result, false)
}

// param: prop 属性
// param: result 填充的struct
// param: withField 是否根据field name填充
// result: result如果不为struct的指针返回错误，填充时异常返回错误
func FillEx(prop Properties, result interface{}, withField bool) error {
	return FillExWithTagName(prop, result, withField, TagPrefixName, TagName)
}

// param: prop 属性
// param: result 填充的struct
// param: withField 是否根据field name填充
// param: tagPxName tag前缀名，后续都使用tagPxName定义的名称做前缀
// param: tagName tag名
// result: result如果不为struct的指针返回错误，填充时异常返回错误
func FillExWithTagName(prop Properties, result interface{}, withField bool, tagPxName, tagName string) error {
	t := reflect.TypeOf(result)
	v := reflect.ValueOf(result)

	if t.Kind() != reflect.Ptr {
		return errors.New("result must be ptr")
	}
	t = t.Elem()
	v = v.Elem()

	if t.Kind() != reflect.Struct {
		return errors.New("result must be struct ptr")
	}

	prefix := ""
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagPxName)
		if tag != "" {
			prefix = tag
			continue
		}
		tag = field.Tag.Get(tagName)
		if tag != "" {
			if tag == "-" {
				continue
			}
		} else if withField {
			tag = field.Name
		}

		if tag != "" {
			if prefix != "" {
				tag = prefix + "." + tag
			}
			c := reflect.New(field.Type).Interface()
			err := prop.GetValue(tag, c)
			if err != nil {
				logf(err.Error())
			}
			fieldValue := v.Field(i)
			if fieldValue.CanSet() {
				fieldValue.Set(reflect.ValueOf(c).Elem())
			}
		}
	}
	return nil
}

// param: prop 属性
// param: result 填充的struct
// param: withField 是否根据field name填充
// param: tagPxNames tag前缀名，后续都使用tagPxName定义的名称做前缀
// param: tagNames tag名
// result: result如果不为struct的指针返回错误，填充时异常返回错误
func FillExWithTagNames(prop Properties, result interface{}, withField bool, tagPxNames, tagNames []string) error {
	if len(tagPxNames) != len(tagNames) {
		return fmt.Errorf("tagPxNames lens not the same with tagNames")
	}
	t := reflect.TypeOf(result)
	v := reflect.ValueOf(result)

	if t.Kind() != reflect.Ptr {
		return errors.New("result must be ptr")
	}
	t = t.Elem()
	v = v.Elem()

	if t.Kind() != reflect.Struct {
		return errors.New("result must be struct ptr")
	}

	errs := Errors{}
	prefix := make([]string, len(tagPxNames))
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		for tagIndex := range tagPxNames {
			tagPxName := tagPxNames[tagIndex]
			tagName := tagNames[tagIndex]
			tagValue := field.Tag.Get(tagPxName)
			if tagValue != "" {
				prefix[tagIndex] = tagValue
				continue
			}
			tagValue = field.Tag.Get(tagName)
			if tagValue != "" {
				if tagValue == "-" {
					break
				}
			} else if tagIndex < len(tagPxNames)-1 {
				continue
			} else if withField {
				tagValue = field.Name
			}

			if tagValue != "" {
				tags := strings.Split(tagValue, ",")
				defaultStr := ""
				if len(tags) > 1 {
					tagValue = tags[0]
					if index := strings.Index(tags[1], "default="); index != -1 {
						defaultStr = tags[1][len("default="):]
					}
				}
				if prefix[tagIndex] != "" {
					tagValue = prefix[tagIndex] + "." + tagValue
				}
				c := reflect.New(field.Type).Interface()
				fieldValue := v.Field(i)
				if defaultStr == "" {
					err := prop.GetValue(tagValue, c)
					if err != nil {
						logf(err.Error())
						errs.AddError(err)
						break
					}
					if fieldValue.CanSet() {
						fieldValue.Set(reflect.ValueOf(c).Elem())
					}
				} else {
					value := prop.Get(tagValue, defaultStr)
					if ok := reflection.SetValue(fieldValue, reflect.ValueOf(value)); !ok {
						errs.AddError(errors.New("Not assigned. "))
					}
				}
				break
			}
		}
	}
	if errs.Empty() {
		return nil
	}
	return errs
}

type Errors []error

func (es Errors) Empty() bool {
	return len(es) == 0
}

func (es *Errors) AddError(e error) *Errors {
	*es = append(*es, e)
	return es
}

func (es Errors) Error() string {
	buf := strings.Builder{}
	for i := range es {
		buf.WriteString(es[i].Error())
		if i < len(es)-1 {
			buf.WriteString(",")
		}
	}
	return buf.String()
}
