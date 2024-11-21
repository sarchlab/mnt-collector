package mntbackend

import (
	"encoding/json"
	"io"
	"reflect"
)

type OKResponse struct {
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func isStructEmpty(data interface{}) bool {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic("isStructEmpty expects a struct or a pointer to a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).IsZero() {
			return false
		}
	}
	return true
}

func unmarshalResponseData(r io.Reader, data interface{}) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var resp OKResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp.Data, data)
	if err != nil {
		return err
	}

	if isStructEmpty(data) {
		return ErrorNilData
	}

	return nil
}

func IsObjectNotFound(err error) bool {
	return err == ObjectNotFound || err == ErrorNilData
}
