package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/model"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var expID primitive.ObjectID

func ExpID() primitive.ObjectID {
	if expID == primitive.NilObjectID {
		log.Panic("exp_id is not initialized")
	}
	return expID
}

func PrepareExpID() {
	expData := model.DBExp{
		ExpKey: model.ExpKey{
			Version: config.C.Experiment.Version,
		},
		Message: config.C.Experiment.Message,
	}
	var err error
	expID, err = GetOrBuildExpID(expData)
	if err != nil {
		log.WithError(err).Fatal("Failed to get exp_id")
	}
	log.WithField("ExpID", expID.Hex()).Info("Successfully get exp_id")
}

func FindExp(data model.ExpKey) (model.DBExp, error) {
	url := fmt.Sprintf("%s/exp/search", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBExp{}, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBExp{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBExp{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.DBExp{}, ErrorStatusNotOK
	}

	var exp model.DBExp
	err = unmarshalResponseData(resp.Body, &exp)
	if err != nil && err != ErrorNilData {
		return model.DBExp{}, err
	}
	if err == ErrorNilData {
		return model.DBExp{}, ObjectNotFound
	}

	return exp, nil
}

func CreateExp(data model.DBExp) (model.DBExp, error) {
	url := fmt.Sprintf("%s/exp", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBExp{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBExp{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBExp{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.DBExp{}, ErrorStatusNotOK
	}

	var exp model.DBExp
	err = unmarshalResponseData(resp.Body, &exp)
	if err != nil {
		return model.DBExp{}, err
	}

	return exp, nil
}

func GetOrBuildExpID(data model.DBExp) (primitive.ObjectID, error) {
	exp, err := FindExp(data.ExpKey)
	if err != nil && err != ObjectNotFound {
		return primitive.NilObjectID, err
	}
	if err == ObjectNotFound {
		exp, err = CreateExp(data)
		if err != nil {
			return primitive.NilObjectID, err
		}
	}

	return exp.ID, nil
}
