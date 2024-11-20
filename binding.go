package ron

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

func (c *CTX) BindJSON(v any) error {
	if c.R.Header.Get("Content-Type") != "application/json" {
		return http.ErrNotSupported
	}
	decoder := json.NewDecoder(c.R.Body)
	return decoder.Decode(v)
}

func (c *CTX) BindForm(v interface{}) error {
	if err := c.R.ParseForm(); err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("v must be a pointer to a struct")
	}

	return mapForm(v, c.R.Form)
}

func mapForm(ptr interface{}, form map[string][]string) error {
	val := reflect.ValueOf(ptr).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := structField.Tag.Get("form")
		if tag == "" {
			tag = structField.Name
		}

		if field.Kind() == reflect.Struct && structField.Anonymous {
			if err := mapForm(field.Addr().Interface(), form); err != nil {
				return err
			}
			continue
		}

		if values, ok := form[tag]; ok && len(values) > 0 {
			if field.Kind() == reflect.Slice {
				elemType := field.Type().Elem()
				slice := reflect.MakeSlice(field.Type(), len(values), len(values))
				for i, v := range values {
					elem := slice.Index(i)
					if elem.Kind() == reflect.Ptr {
						elem.Set(reflect.New(elemType.Elem()))
						elem = elem.Elem()
					}
					if err := setField(elem, v); err != nil {
						return err
					}
				}
				field.Set(slice)
			} else {
				if err := setField(field, values[0]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func setField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return nil
	}

	if field.Type() == reflect.TypeOf(time.Time{}) {
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(t))
		return nil
	}

	kind := field.Kind()
	switch kind {
	case reflect.String:
		field.SetString(value)
	case reflect.Int:
		intValue, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Uint:
		uintValue, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	default:
		return fmt.Errorf("unsupported type: %s", kind)
	}
	return nil
}
