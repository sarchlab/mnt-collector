package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sarchlab/mnt-collector/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EnvRequest struct {
	GPU         string `json:"gpu"`
	Machine     string `json:"machine"`
	CUDAVersion string `json:"cuda_version"`
}

func GetEnvID(data EnvRequest) (primitive.ObjectID, error) {
	url := fmt.Sprintf("%s/env-id", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return primitive.NilObjectID, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return primitive.NilObjectID, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return primitive.NilObjectID, ErrorStatusNotOK
	}

	var envID primitive.ObjectID
	err = unmarshalResponseData(resp.Body, &envID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return envID, nil
}
