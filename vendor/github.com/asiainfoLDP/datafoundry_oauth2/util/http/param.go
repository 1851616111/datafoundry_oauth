package http

import (
	"bytes"
	"fmt"
	"reflect"
)

type Param interface {
	GetBodyType() string
	GetParam() interface{}
}

func InterfaceToString(param interface{}) string {
	v := reflect.ValueOf(param)
	t := v.Type()

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	fields := t.NumField()
	var buf bytes.Buffer

	for i := 0; i < fields; i++ {
		tag := t.Field(i).Tag.Get("param")
		if tag == "-" {
			continue
		}

		if len(v.Field(i).String()) > 0 {
			if buf.Len() > 0 {
				buf.WriteString(`&`)
			}

			if len(tag) > 0 {
				buf.WriteString(tag)
			} else {
				buf.WriteString(t.Field(i).Name)
			}

			buf.WriteString("=")
			switch t.Field(i).Type.Kind() {
			case reflect.Bool:
				buf.WriteString(fmt.Sprintf("%t", v.Field(i).Bool()))
			case reflect.String:
				buf.WriteString(v.Field(i).String())
			case reflect.Float32, reflect.Float64:
				buf.WriteString(fmt.Sprintf("%f", v.Field(i).Float()))
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				buf.WriteString(fmt.Sprintf("%d", v.Field(i).Int()))
			}
		}
	}

	return buf.String()
}
