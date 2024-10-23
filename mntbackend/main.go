package mntbackend

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/sarchlab/mnt-collector/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var URLBase string
var envID primitive.ObjectID

func EnvID() primitive.ObjectID {
	return envID
}

func Init() {
	c := config.SC.MNT
	URLBase = fmt.Sprintf("http://%s:%d%s", c.Host, c.Port, c.Base)

	err := checkHealth()
	if err != nil {
		log.Fatal(err)
	}

	envData := EnvRequest{
		GPU:         config.DeviceName(),
		Machine:     config.HostName(),
		CUDAVersion: config.CudaVersion(),
	}
	envID, err = GetEnvID(envData)
	if err != nil {
		log.Panic(err)
	}
	log.WithField("EnvID", envID.Hex()).Info("Successfully get env_id")
}
