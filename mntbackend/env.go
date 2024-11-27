package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sarchlab/mnt-collector/config"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sarchlab/mnt-backend/model"
)

var envID primitive.ObjectID

func EnvID() primitive.ObjectID {
	if envID == primitive.NilObjectID {
		log.Panic("env_id is not initialized")
	}
	return envID
}

func PrepareEnvID() {
	envData := model.DBEnv{
		EnvKey: model.EnvKey{
			GPU:         config.DeviceName(),
			Machine:     config.HostName(),
			CUDAVersion: config.CudaVersion(),
		},
	}
	var err error
	envID, err = GetOrBuildEnvID(envData)
	if err != nil {
		log.WithError(err).Panic("Failed to get env_id")
	}
	log.WithField("EnvID", envID.Hex()).Info("Successfully get env_id")
}

func FindEnv(data model.EnvKey) (model.DBEnv, error) {
	url := fmt.Sprintf("%s/env/search", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBEnv{}, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBEnv{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBEnv{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.DBEnv{}, ErrorStatusNotOK
	}

	var env model.DBEnv
	err = unmarshalResponseData(resp.Body, &env)
	if err != nil && err != ErrorNilData {
		return model.DBEnv{}, err
	}
	if err == ErrorNilData {
		return model.DBEnv{}, ObjectNotFound
	}

	return env, nil
}

func CreateEnv(data model.DBEnv) (model.DBEnv, error) {
	url := fmt.Sprintf("%s/env", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBEnv{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBEnv{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBEnv{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.DBEnv{}, ErrorStatusNotOK
	}

	var env model.DBEnv
	err = unmarshalResponseData(resp.Body, &env)
	if err != nil {
		return model.DBEnv{}, err
	}

	return env, nil
}

func GetOrBuildEnvID(data model.DBEnv) (primitive.ObjectID, error) {
	env, err := FindEnv(data.EnvKey)
	if err != nil && err != ObjectNotFound {
		return primitive.NilObjectID, err
	}
	if err == ObjectNotFound {
		env, err = CreateEnv(data)
		if err != nil {
			return primitive.NilObjectID, err
		}
	}

	return env.ID, nil
}
