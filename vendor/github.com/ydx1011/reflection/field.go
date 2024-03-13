package reflection

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func SetStrcutFieldValue(o interface{}, fieldStr string, value interface{}) error {
	return SetStrcutFieldValueByTag(o, fieldStr, value, "")
}

func SetStrcutFieldValueByTag(o interface{}, matchStr string, value interface{}, tagName string) error {
	return SetStrcutFieldValueEx(o, matchStr, value, tagName, nil)
}

func SetStrcutFieldValueEx(o interface{}, fieldName string, value interface{}, tagName string, modifier func(string) string) error {
	return SetFieldValueEx(reflect.ValueOf(o), fieldName, reflect.ValueOf(value), tagName, modifier)
}

func SetFieldValue(dst reflect.Value, fieldName string, value reflect.Value) error {
	return SetFieldValueByTag(dst, fieldName, value, "")
}

func SetFieldValueByTag(dst reflect.Value, tags string, value reflect.Value, tagName string) error {
	return SetFieldValueEx(dst, tags, value, tagName, nil)
}

func SetFieldValueEx(v reflect.Value, fieldName string, value reflect.Value, tagName string, modifier func(string) string) error {
	t := v.Type()
	if t.Kind() != reflect.Ptr {
		return errors.New("Set dest object must be struct pointer. ")
	}
	t = t.Elem()
	v = v.Elem()
	if t.Kind() != reflect.Struct {
		return errors.New("Set dest object must be struct pointer. ")
	}

	fields := strings.Split(fieldName, ".")
	fv := v
	ft := t
	for i, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if modifier != nil {
			f = modifier(f)
		}
		if tagName != "" {
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
				fv = fv.Elem()
			}
			fs := fv.NumField()
			found := false
			for j := 0; j < fs; j++ {
				ff := ft.Field(j)
				if tag, ok := ff.Tag.Lookup(tagName); ok {
					if tag == f {
						ft = ff.Type
						fv = fv.Field(j)

						found = true
						break
					}
				}
			}
			if !found {
				return fmt.Errorf("Field: %s at %d not found. ", f, i)
			}
		} else {
			if fv.Kind() == reflect.Ptr {
				fv = fv.Elem()
			}
			fv = fv.FieldByName(f)
			if !fv.IsValid() {
				return fmt.Errorf("Field: %s at %d not found. ", f, i)
			}
		}
	}

	if SetValue(fv, value) {
		return nil
	} else {
		return fmt.Errorf("Value type is not assiginable to field. ")
	}
}
