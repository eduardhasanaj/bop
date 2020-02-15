package bop

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var (
	urlEncDelimiter          = []byte("&")
	eqDelimiter              = []byte("=")
	quote                    = []byte(`"`)
	maxBodyPayloadSize int64 = 1024 * 1024

	bindingMaps map[reflect.Type]map[string]int = make(map[reflect.Type]map[string]int, 0)
	mut         sync.RWMutex
)

type BodyParser struct {
	Form url.Values
	r    *http.Request
	w    http.ResponseWriter
}

func New(w http.ResponseWriter, r *http.Request) *BodyParser {
	return &BodyParser{
		r: r,
		w: w,
	}
}

//Parse body into a model
//Returns parsed property names
func (bp *BodyParser) ParseModel(model interface{}) ([]string, error) {
	//Body Parser by design operate only for POST, PUT, PATCH
	if bp.r.Method != "POST" && bp.r.Method != "PUT" && bp.r.Method != "PATCH" {
		return nil, errors.New("BodyParser supports only http methods of POST, PUT, PATCH")
	}

	t := reflect.ValueOf(model).Elem()
	if t.Kind() != reflect.Struct {
		return nil, errors.New("Only models of kind Struct are supported")
	}

	mut.RLock()
	bindingMap, ok := bindingMaps[t.Type()]
	if !ok {
		bindingMap = makeBindingMap(&t)
		bindingMaps[t.Type()] = bindingMap
	}
	mut.RUnlock()

	//Limit body size
	bp.r.Body = http.MaxBytesReader(bp.w, bp.r.Body, maxBodyPayloadSize)

	ct := bp.r.Header.Get("Content-Type")

	switch ct {
	case "application/json":
		columns, err := parseJsonModelUsingIterator(bp.r, &t, bindingMap)
		return columns, err
	case "application/x-www-form-urlencoded", "multipart/form-data":
		columns, err := parseFromPostForm(bp.r, ct, &t, bindingMap)
		return columns, err
	}

	return nil, nil
}

func parseJsonModel(r *http.Request, t *reflect.Value, bindingMap map[string]int) ([]string, error) {
	dec := jsoniter.NewDecoder(r.Body)

	var json map[string]jsoniter.RawMessage

	if err := dec.Decode(&json); err != nil {
		return nil, err
	}

	columns := make([]string, 0, len(bindingMap))

	for key, v := range json {
		fIndex, ok := bindingMap[key]
		if !ok {
			return nil, errors.New("could not bind to model: key " + key)
		}

		if err := jsoniter.Unmarshal(v, t.Field(fIndex).Addr().Interface()); err != nil {
			return nil, errors.New("could not bind to model: key " + key)
		}

		columns = append(columns, key)
	}

	return columns, nil
}

func parseJsonModelUsingIterator(r *http.Request, t *reflect.Value, bindingMap map[string]int) ([]string, error) {
	it := jsoniter.Parse(jsoniter.ConfigDefault, r.Body, 64)
	if it.Error != nil {
		return nil, it.Error
	}

	columns := make([]string, 0, len(bindingMap))
	key := it.ReadObject()
	if it.Error != nil {
		return nil, it.Error
	}

	for key != "" {
		valType := it.WhatIsNext()
		if valType == jsoniter.ObjectValue || valType == jsoniter.ArrayValue {
			return nil, errors.New("ParseModel supports only flat object model")
		}

		fIndex, ok := bindingMap[key]
		if !ok {
			return nil, errors.New("could not bind to model: key " + key)
		}

		it.ReadVal(t.Field(fIndex).Addr().Interface())
		if it.Error != nil {
			return nil, errors.New("could not bind to model")
		}

		columns = append(columns, key)

		key = it.ReadObject()
		if it.Error != nil {
			return nil, it.Error
		}
	}

	return columns, nil
}

func parseFromPostForm(r *http.Request, ct string, t *reflect.Value, bindingMap map[string]int) ([]string, error) {
	var err error
	if ct == "multipart/form-data" {
		err = r.ParseMultipartForm(maxBodyPayloadSize)
	} else {
		err = r.ParseForm()
	}

	if err != nil {
		return nil, err
	}

	columns := make([]string, 0, len(bindingMap))

	for k, v := range r.PostForm {
		fIndex, ok := bindingMap[k]
		if !ok {
			return nil, errors.New("could not bind to model")
		}
		val := t.Field(fIndex)
		valBytes := []byte(v[0])

		//In case of string quote
		if val.Kind() == reflect.String {
			valBytes = quoteString(valBytes, quote)
		}

		if err := jsoniter.Unmarshal(valBytes, val.Addr().Interface()); err != nil {
			return nil, errors.New("could not bind to model")
		}

		columns = append(columns, k)
	}

	return columns, nil
}

func makeBindingMap(t *reflect.Value) map[string]int {
	bindingMap := make(map[string]int)
	for i := 0; i != t.NumField(); i++ {

		if t.Field(i).CanSet() {
			key := t.Type().Field(i).Tag.Get("json")
			bindingMap[key] = i
		}
	}

	return bindingMap
}

func quoteString(s []byte, q []byte) []byte {
	buff := make([]byte, 0, len(s)+len(q)*2)
	buff = append(buff, q...)
	buff = append(buff, s...)
	buff = append(buff, q...)

	return buff
}
