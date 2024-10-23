package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sarchlab/mnt-collector/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileRequest struct {
	EnvID     primitive.ObjectID `json:"env_id" bson:"env_id"`
	Suite     string             `json:"suite" bson:"suite"`
	Benchmark string             `json:"benchmark" bson:"benchmark"`
	Param     Param              `json:"param" bson:"param"`

	RepeatTimes int32   `json:"repeat_times" bson:"repeat_times"`
	AvgCycles   float64 `json:"avg_cycles" bson:"avg_cycles"`
}

type Param struct {
	Size       int32 `json:"size,omitempty" bson:"size,omitempty"`
	VectorN    int32 `json:"vectorN,omitempty" bson:"vectorN,omitempty"`
	ElementN   int32 `json:"elementN,omitempty" bson:"elementN,omitempty"`
	Log2Data   int32 `json:"log2data,omitempty" bson:"log2data,omitempty"`
	Log2Kernel int32 `json:"log2kernel,omitempty" bson:"log2kernel,omitempty"`
}

func UploadProfile(data ProfileRequest) (primitive.ObjectID, error) {
	url := fmt.Sprintf("%s/profile", URLBase)

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

	var profileID primitive.ObjectID
	err = unmarshalResponseData(resp.Body, &profileID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return profileID, nil
}
