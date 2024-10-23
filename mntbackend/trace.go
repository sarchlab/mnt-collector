package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sarchlab/mnt-collector/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TraceRequest struct {
	EnvID     primitive.ObjectID `bson:"env_id" json:"env_id"`
	Suite     string             `bson:"suite" json:"suite"`
	Benchmark string             `bson:"benchmark" json:"benchmark"`
	Param     Param              `bson:"param" json:"param"`

	S3Path string `bson:"s3_path" json:"s3_path"`
}

func UploadTrace(data TraceRequest) (primitive.ObjectID, error) {
	url := fmt.Sprintf("%s/trace", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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

	var traceID primitive.ObjectID
	err = unmarshalResponseData(resp.Body, &traceID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return traceID, nil
}
