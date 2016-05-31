package http

import (
	"bytes"
	"reflect"
)

const (
	ContentType_Form = "application/x-www-form-urlencoded"
	ContentType_Json = "application/json"
)

type Param interface {
	GetBodyType() string
	GetParam() interface{}
}

type NewOption struct {
	Param            interface{}
	ParamContentType string
}

func (p *NewOption) GetBodyType() string {
	return p.ParamContentType
}

func (p *NewOption) GetParam() interface{} {
	return p.Param
}

func InterfaceToString(param interface{}) string {
	v := reflect.ValueOf(param)
	t := v.Type()

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	fields := t.NumField()
	strType := reflect.TypeOf("")

	var buf bytes.Buffer

	for i := 0; i < fields; i++ {
		if t.Field(i).Type != strType {
			continue
		}

		if v.Field(i).Len() > 0 {
			if buf.Len() > 0 {
				buf.WriteString("&")
			}

			tag := t.Field(i).Tag.Get("param")
			if len(tag) > 0 {
				buf.WriteString(tag)
			} else {
				buf.WriteString(t.Field(i).Name)
			}

			buf.WriteString("=")
			buf.WriteString(v.Field(i).String())
		}
	}

	return buf.String()
}
