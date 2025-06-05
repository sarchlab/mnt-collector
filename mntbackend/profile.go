package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateProfile(data model.DBProf) (model.DBProf, error) {
	url := fmt.Sprintf("%s/profile", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBProf{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBProf{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBProf{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.DBProf{}, ErrorStatusNotOK
	}

	var profile model.DBProf
	err = unmarshalResponseData(resp.Body, &profile)
	if err != nil {
		return model.DBProf{}, err
	}

	return profile, nil
}

func UpdateProfile(id primitive.ObjectID, data model.DBProf) error {
	url := fmt.Sprintf("%s/profile/%s", URLBase, id.Hex())

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
		return ErrorStatusNotOK
	}

	return nil
}

func FindProfile(data model.CaseKey) (model.DBProf, error) {
	url := fmt.Sprintf("%s/profile/search", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBProf{}, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBProf{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBProf{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.DBProf{}, ErrorStatusNotOK
	}

	var profile model.DBProf
	err = unmarshalResponseData(resp.Body, &profile)
	if err != nil && err != ErrorNilData {
		return model.DBProf{}, err
	}
	if err == ErrorNilData {
		return model.DBProf{}, ObjectNotFound
	}

	return profile, nil
}

func UpdOrUplProfile(data model.DBProf) (primitive.ObjectID, error) {
	profile, err := FindProfile(data.CaseKey)
	if err != nil {
		if err == ObjectNotFound {
			createdProfile, err := CreateProfile(data)
			if err != nil {
				return primitive.NilObjectID, err
			}
			return createdProfile.ID, nil
		}
		return primitive.NilObjectID, err
	}

	err = UpdateProfile(profile.ID, data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return profile.ID, nil
}
