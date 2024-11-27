package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sarchlab/mnt-backend/model"
	"github.com/sarchlab/mnt-collector/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateSimulation(data model.DBSim) (model.DBSim, error) {
	url := fmt.Sprintf("%s/sim", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBSim{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBSim{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBSim{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.DBSim{}, ErrorStatusNotOK
	}

	var sim model.DBSim
	err = unmarshalResponseData(resp.Body, &sim)
	if err != nil {
		return model.DBSim{}, err
	}

	return sim, nil
}

func UpdateSimulation(id primitive.ObjectID, data model.DBSim) error {
	url := fmt.Sprintf("%s/sim/%s", URLBase, id.Hex())

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

func FindSim(data model.SimKey) (model.DBSim, error) {
	url := fmt.Sprintf("%s/sim/search", URLBase)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBSim{}, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return model.DBSim{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.DBSim{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.DBSim{}, ErrorStatusNotOK
	}

	var sim model.DBSim
	err = unmarshalResponseData(resp.Body, &sim)
	if err != nil && err != ErrorNilData {
		return model.DBSim{}, err
	}
	if err == ErrorNilData {
		return model.DBSim{}, ObjectNotFound
	}

	return sim, nil
}

func UpdOrUplSim(data model.DBSim) (primitive.ObjectID, error) {
	sim, err := FindSim(data.SimKey)
	if err != nil {
		if err == ObjectNotFound {
			sim, err = CreateSimulation(data)
			if err != nil {
				return primitive.NilObjectID, err
			}
			return sim.ID, nil
		}
		return primitive.NilObjectID, err
	}

	err = UpdateSimulation(sim.ID, data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return sim.ID, nil
}
