package mntbackend

import (
	"encoding/json"
	"io"
)

type OKResponse struct {
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
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

	return nil
}
