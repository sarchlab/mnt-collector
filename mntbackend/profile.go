package mntbackend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// func CreateProfile(data model.DBProf) (model.DBProf, error) {
// 	url := fmt.Sprintf("%s/profile", URLBase)

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		return model.DBProf{}, err
// 	}

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return model.DBProf{}, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return model.DBProf{}, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return model.DBProf{}, ErrorStatusNotOK
// 	}

// 	var profile model.DBProf
// 	err = unmarshalResponseData(resp.Body, &profile)
// 	if err != nil {
// 		return model.DBProf{}, err
// 	}

// 	return profile, nil
// }

// func UpdateProfile(id primitive.ObjectID, data model.DBProf) error {
// 	url := fmt.Sprintf("%s/profile/%s", URLBase, id.Hex())

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		return err
// 	}

// 	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return ErrorStatusNotOK
// 	}

// 	return nil
// }

// func FindProfile(data model.CaseKeyProfile) (model.DBProf, error) {
// 	url := fmt.Sprintf("%s/profile/search", URLBase)

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		return model.DBProf{}, err
// 	}

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return model.DBProf{}, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SC.MNT.Token))

// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return model.DBProf{}, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return model.DBProf{}, ErrorStatusNotOK
// 	}

// 	var profile model.DBProf
// 	err = unmarshalResponseData(resp.Body, &profile)
// 	if err != nil && err != ErrorNilData {
// 		return model.DBProf{}, err
// 	}
// 	if err == ErrorNilData {
// 		return model.DBProf{}, ObjectNotFound
// 	}

// 	return profile, nil
// }

// func UpdOrUplProfile(data model.DBProf) (primitive.ObjectID, error) {
// 	found, err := FindProfile(data.CaseKeyProfile)
// 	fmt.Printf("data.CaseKeyProfile: %+v\n", data.CaseKeyProfile)
// 	if err != nil {
// 		if err == ObjectNotFound {
// 			createdProfile, err := CreateProfile(data)
// 			if err != nil {
// 				return primitive.NilObjectID, err
// 			}
// 			return createdProfile.ID, nil
// 		}
// 		return primitive.NilObjectID, err
// 	}

// 	// err = UpdateProfile(profile.ID, data)
// 	// if err != nil {
// 	// 	return primitive.NilObjectID, err
// 	// }

// 	// return profile.ID, nil
// 	// Safety: if server-side search ignored profile_type and returned another type, create a new one.
// 	if found.ProfileType != data.ProfileType {
// 		created, err := CreateProfile(data)
// 		if err != nil {
// 			return primitive.NilObjectID, err
// 		}
// 		return created.ID, nil
// 	}

// 	if err := UpdateProfile(found.ID, data); err != nil {
// 		return primitive.NilObjectID, err
// 	}
// 	return found.ID, nil
// }

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
		body, _ := io.ReadAll(resp.Body)
		return model.DBProf{}, fmt.Errorf("create failed status=%d body=%s", resp.StatusCode, body)
	}
	var profile model.DBProf
	if err := unmarshalResponseData(resp.Body, &profile); err != nil {
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed status=%d body=%s", resp.StatusCode, body)
	}
	return nil
}

func FindProfile(data model.CaseKeyProfile) (model.DBProf, error) {
	url := fmt.Sprintf("%s/profile/search", URLBase)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return model.DBProf{}, err
	}
	// Use POST (body honored)
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
		if resp.StatusCode == http.StatusNotFound {
			return model.DBProf{}, ObjectNotFound
		}
		body, _ := io.ReadAll(resp.Body)
		return model.DBProf{}, fmt.Errorf("search failed status=%d body=%s", resp.StatusCode, body)
	}
	var profile model.DBProf
	err = unmarshalResponseData(resp.Body, &profile)
	if err != nil && err != ErrorNilData {
		return model.DBProf{}, err
	}
	if err == ErrorNilData || profile.ID.IsZero() {
		return model.DBProf{}, ObjectNotFound
	}
	return profile, nil
}

func UpdOrUplProfile(data model.DBProf) (primitive.ObjectID, error) {
	fmt.Printf("DEBUG CaseKeyProfile: %+v\n", data.CaseKeyProfile)
	found, err := FindProfile(data.CaseKeyProfile)
	if err != nil {
		if err == ObjectNotFound {
			created, cerr := CreateProfile(data)
			if cerr != nil {
				return primitive.NilObjectID, cerr
			}
			return created.ID, nil
		}
		return primitive.NilObjectID, err
	}
	// If somehow server returned a different profile_type document, force create
	if found.ProfileType != data.ProfileType {
		created, cerr := CreateProfile(data)
		if cerr != nil {
			return primitive.NilObjectID, cerr
		}
		return created.ID, nil
	}
	if uerr := UpdateProfile(found.ID, data); uerr != nil {
		return primitive.NilObjectID, uerr
	}
	return found.ID, nil
}
