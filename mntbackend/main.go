package mntbackend

import (
	"encoding/json"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/sarchlab/mnt-collector/collector"
	"github.com/sarchlab/mnt-collector/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var URLBase string
var EnvID primitive.ObjectID

func Init() {
	c := config.SC.MNT
	URLBase = fmt.Sprintf("http://%s:%d%s", c.Host, c.Port, c.Base)

	err := checkHealth()
	if err != nil {
		log.Fatal(err)
	}

	envData := EnvRequest{
		GPU:         collector.DeviceName(),
		Machine:     config.HostName(),
		CUDAVersion: config.CudaVersion(),
	}
	EnvID, err = GetEnvID(envData)
	if err != nil {
		log.Panic(err)
	}
	log.WithField("EnvID", EnvID.Hex()).Info("Successfully get env_id")
}

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
