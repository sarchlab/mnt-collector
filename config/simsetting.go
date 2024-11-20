package config

import (
	"os"
	"path/filepath"

	"github.com/sarchlab/mnt-backend/model"
	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
)

type Case struct {
	Title     string        `yaml:"title"`
	Suite     string        `yaml:"suite"`
	Directory string        `yaml:"directory"`
	Command   string        `yaml:"command"`
	Args      []model.Param `yaml:"args"`
}

type SimSetting struct {
	DeviceID      int  `yaml:"device-id"`
	ExclusiveMode bool `yaml:"exclusive-mode"`

	UploadToServer bool `yaml:"upload-to-server"`
	TraceCollect   struct {
		Enable bool `yaml:"enable"`
	} `yaml:"trace-collect"`
	ProfileCollect struct {
		Enable      bool  `yaml:"enable"`
		RepeatTimes int32 `yaml:"repeat-times"`
	} `yaml:"profile-collect"`

	Cases []Case `yaml:"cases"`
}

func (c *SimSetting) load(file string) {
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
