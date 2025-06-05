package config

import (
	"os"
	"path/filepath"

	"github.com/sarchlab/mnt-collector/externel/mnt-backend/model"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gopkg.in/yaml.v3"
)

type CollectionSettings struct {
	Experiment struct {
		Version string `yaml:"version"`
		Message string `yaml:"message"`
		Runfile string `yaml:"runfile"`
	} `yaml:"experiment"` // simulations
	TraceIDs []primitive.ObjectID `yaml:"trace-id"` // simulations

	RepeatTimes int32 `yaml:"repeat-times"` // profiles

	DeviceID       int    `yaml:"device-id"`        // traces & profiles
	ExclusiveMode  bool   `yaml:"exclusive-mode"`   // traces & profiles
	UploadToServer bool   `yaml:"upload-to-server"` // traces & profiles
	Cases          []Case `yaml:"cases"`            // traces & profiles
}

type Case struct {
	Title     string        `yaml:"title"`
	Suite     string        `yaml:"suite"`
	Directory string        `yaml:"directory"`
	Command   string        `yaml:"command"`
	Args      []model.Param `yaml:"args"`
}

func (c *CollectionSettings) load(file string) {
	file = filepath.Join(projectRoot, file)
	bytes, err := os.ReadFile(file)
	if err != nil {
		log.WithError(err).WithField("file", file).Panic("Failed to read file")
	}
	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		log.WithError(err).WithField("file", file).Panic("Failed to unmarshal yaml")
	}
}
