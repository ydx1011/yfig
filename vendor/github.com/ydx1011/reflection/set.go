package reflection

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func SetValue(dst reflect.Value, value reflect.Value) bool {
	hasAssigned := false
	vt := value.Type()

	dt := dst.Type()
	switch dt.Kind() {
	case reflect.Bool:
		switch vt.Kind() {
		case reflect.Bool:
			hasAssigned = true
			dst.SetBool(value.Bool())
			break
		case reflect.Slice:
			if d, ok := value.Interface().([]uint8); ok {
				hasAssigned = true
				dst.SetBool(d[0] != 0)
			}
			break
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			hasAssigned = true
			dst.SetBool(value.Uint() != 0)
			break
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			hasAssigned = true
			dst.SetBool(value.Int() != 0)
			break
		case reflect.String:
			b, err := strconv.ParseBool(value.String())
			if err == nil {
				hasAssigned = true
				dst.SetBool(b)
			}
			break
		}
		break
	case reflect.Map:
		switch vt.Kind() {
		case reflect.Map:
			_, err := SetOrCopyMap(dst, value, true)
			hasAssigned = err == nil
		}
		break
	case reflect.Slice:
		switch vt.Kind() {
		case reflect.String:
			if dt.Elem().Kind() == reflect.Uint8 {
				hasAssigned = true
				dst.SetBytes([]byte(value.String()))
			}
			break
		case reflect.Slice:
			_, err := SetOrCopySlice(dst, value, true)
			hasAssigned = err == nil
			break
		}
	case reflect.String:
		switch vt.Kind() {
		case reflect.String:
			hasAssigned = true
			dst.SetString(value.String())
			break
		case reflect.Slice:
			if d, ok := value.Interface().([]uint8); ok {
				hasAssigned = true
				dst.SetString(string(d))
			}
			break
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			hasAssigned = true
			dst.SetString(strconv.FormatUint(value.Uint(), 10))
			break
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			hasAssigned = true
			dst.SetString(strconv.FormatInt(value.Int(), 10))
			break
		case reflect.Float64:
			hasAssigned = true
			dst.SetString(strconv.FormatFloat(value.Float(), 'g', -1, 64))
			break
		case reflect.Float32:
			hasAssigned = true
			dst.SetString(strconv.FormatFloat(value.Float(), 'g', -1, 32))
			break
		case reflect.Bool:
			hasAssigned = true
			dst.SetString(strconv.FormatBool(value.Bool()))
			break
		//case reflect.Struct:
		//    if ti, ok := v.(time.Time); ok {
		//        hasAssigned = true
		//        if ti.IsZero() {
		//            dst.SetString("")
		//        } else {
		//            dst.SetString(ti.String())
		//        }
		//    } else {
		//        hasAssigned = true
		//        dst.SetString(fmt.Sprintf("%v", v))
		//    }
		default:
			hasAssigned = true
			dst.SetString(fmt.Sprintf("%v", value.Interface()))
		}
		break
	case reflect.Complex64, reflect.Complex128:
		switch vt.Kind() {
		case reflect.Complex64, reflect.Complex128:
			hasAssigned = true
			dst.SetComplex(value.Complex())
			break
		case reflect.Slice:
			if vt.ConvertibleTo(BytesType) {
				d := value.Bytes()
				if len(d) > 0 {
					if dst.CanAddr() {
						err := json.Unmarshal(d, dst.Addr().Interface())
						if err != nil {
							return false
						}
					} else {
						x := reflect.New(dt)
						err := json.Unmarshal(d, x.Interface())
						if err != nil {
							return false
						}
						hasAssigned = true
						dst.Set(x.Elem())
						break
					}
				}
			}
			break
		}
		break
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch vt.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			hasAssigned = true
			dst.SetInt(value.Int())
			break
		case reflect.Slice:
			if d, ok := value.Interface().([]uint8); ok {
				intV, err := strconv.ParseInt(string(d), 10, 64)
				if err == nil {
					hasAssigned = true
					dst.SetInt(intV)
				}
			}
			break
		case reflect.String:
			b, err := strconv.ParseInt(value.String(), 10, 64)
			if err == nil {
				hasAssigned = true
				dst.SetInt(b)
			}
			break
		}
		break
	case reflect.Float32, reflect.Float64:
		switch vt.Kind() {
		case reflect.Float32, reflect.Float64:
			hasAssigned = true
			dst.SetFloat(value.Float())
			break
		case reflect.Slice:
			if d, ok := value.Interface().([]uint8); ok {
				floatV, err := strconv.ParseFloat(string(d), 64)
				if err == nil {
					hasAssigned = true
					dst.SetFloat(floatV)
				}
			}
			break
		case reflect.String:
			b, err := strconv.ParseFloat(value.String(), 10)
			if err == nil {
				hasAssigned = true
				dst.SetFloat(b)
			}
			break
		}
		break
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		switch vt.Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			hasAssigned = true
			dst.SetUint(value.Uint())
			break
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			hasAssigned = true
			dst.SetUint(uint64(value.Int()))
			break
		case reflect.Slice:
			if d, ok := value.Interface().([]uint8); ok {
				uintV, err := strconv.ParseUint(string(d), 10, 64)
				if err == nil {
					hasAssigned = true
					dst.SetUint(uintV)
				}
			}
			break
		case reflect.String:
			b, err := strconv.ParseUint(value.String(), 10, 64)
			if err == nil {
				hasAssigned = true
				dst.SetUint(b)
			}
			break
		}
		break
	case reflect.Struct:
		fieldType := dst.Type()
		if fieldType.ConvertibleTo(TimeType) {
			if vt == TimeType {
				hasAssigned = true
				t := value.Convert(TimeType).Interface().(time.Time)
				dst.Set(reflect.ValueOf(t).Convert(fieldType))
			} else if vt == IntType || vt == Int64Type ||
				vt == Int32Type {
				hasAssigned = true

				t := time.Unix(value.Int(), 0)
				dst.Set(reflect.ValueOf(t).Convert(fieldType))
			} else if vt == StringType {
				t, err := convert2Time([]byte(value.String()), time.Local)
				if err == nil {
					hasAssigned = true
					dst.Set(reflect.ValueOf(t).Convert(fieldType))
				}
			} else {
				if d, ok := value.Interface().([]byte); ok {
					t, err := convert2Time(d, time.Local)
					if err == nil {
						hasAssigned = true
						dst.Set(reflect.ValueOf(t).Convert(fieldType))
					}
				}
			}
		} else {
			if vt.AssignableTo(dt) {
				hasAssigned = true
				dst.Set(value)
			} else if vt.ConvertibleTo(dt) {
				hasAssigned = true
				dst.Set(value.Convert(dt))
			}
		}
		break
	case reflect.Ptr:
		if vt.Kind() == reflect.Ptr {
			if vt.AssignableTo(dt) {
				hasAssigned = true
				dst.Set(value)
			} else if vt.ConvertibleTo(dt) {
				hasAssigned = true
				dst.Set(value.Convert(dt))
			}
		}
		break
	case reflect.Chan:
		if vt.Kind() == reflect.Chan {
			if vt.AssignableTo(dt) {
				hasAssigned = true
				dst.Set(value)
			} else if vt.ConvertibleTo(dt) {
				hasAssigned = true
				dst.Set(value.Convert(dt))
			}
		}
		break
	case reflect.Interface:
		hasAssigned = true
		dst.Set(value)
		break
	}

	return hasAssigned
}

const (
	zeroTime0 = "0000-00-00 00:00:00"
	zeroTime1 = "0001-01-01 00:00:00"
)

func convert2Time(data []byte, location *time.Location) (time.Time, error) {
	timeStr := strings.TrimSpace(string(data))
	var timeRet time.Time
	var err error
	if timeStr == zeroTime0 || timeStr == zeroTime1 {
	} else if !strings.ContainsAny(timeStr, "- :") {
		// time stamp
		sd, err := strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			timeRet = time.Unix(sd, 0)
		}
	} else if len(timeStr) > 19 && strings.Contains(timeStr, "-") {
		timeRet, err = time.ParseInLocation(time.RFC3339Nano, timeStr, location)
		if err != nil {
			timeRet, err = time.ParseInLocation("2006-01-02 15:04:05.999999999", timeStr, location)
		}
		if err != nil {
			timeRet, err = time.ParseInLocation("2006-01-02 15:04:05.9999999 Z07:00", timeStr, location)
		}
	} else if len(timeStr) == 19 && strings.Contains(timeStr, "-") {
		timeRet, err = time.ParseInLocation("2006-01-02 15:04:05", timeStr, location)
	} else if len(timeStr) == 10 && timeStr[4] == '-' && timeStr[7] == '-' {
		timeRet, err = time.ParseInLocation("2006-01-02", timeStr, location)
	}
	return timeRet, nil
}
