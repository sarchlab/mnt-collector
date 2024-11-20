package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sarchlab/mnt-backend/model"
	"github.com/sarchlab/mnt-collector/config"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateTrace(data model.DBTrace) (model.DBTrace, error) {
	url := fmt.Sprintf("%s/trace", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBTrace{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBTrace{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBTrace{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status", resp.Status).Warn("failed to create trace")
		return model.DBTrace{}, ErrorStatusNotOK
	}

	var trace model.DBTrace
	err = unmarshalResponseData(resp.Body, &trace)
	if err != nil {
		return model.DBTrace{}, err
	}

	return trace, nil
}

func UpdateTrace(id primitive.ObjectID, data model.DBTrace) error {
	url := fmt.Sprintf("%s/trace/%s", URLBase, id.Hex())

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status", resp.Status).Warn("failed to update trace")
		return ErrorStatusNotOK
	}

	var trace model.DBTrace
	err = unmarshalResponseData(resp.Body, &trace)
	if err != nil {
		return err
	}

	return nil
}

func FindTrace(data model.CaseKey) (model.DBTrace, error) {
	url := fmt.Sprintf("%s/trace/search", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBTrace{}, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBTrace{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBTrace{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status", resp.Status).Warn("failed to find trace")
		return model.DBTrace{}, ErrorStatusNotOK
	}

	var trace model.DBTrace
	err = unmarshalResponseData(resp.Body, &trace)
	if err != nil && err != ErrorNilData {
		return model.DBTrace{}, err
	}
	if err == ErrorNilData {
		return model.DBTrace{}, ObjectNotFound
	}

	return trace, nil
}

func UpdOrUplTrace(data model.DBTrace) (primitive.ObjectID, error) {
	trace, err := FindTrace(data.CaseKey)
	if err != nil {
		if err == ObjectNotFound {
			trace, err = CreateTrace(data)
			if err != nil {
				return primitive.NilObjectID, err
			}
			return trace.ID, nil
		}
		return primitive.NilObjectID, err
	}

	err = UpdateTrace(trace.ID, data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return trace.ID, nil
}
