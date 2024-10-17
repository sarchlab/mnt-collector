package config

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
)

type SimSetting struct {
	DeviceID int  `yaml:"device-id"`
	Root     bool `yaml:"root"`

	UploadToServer bool `yaml:"upload-to-server"`
	TraceCollect   struct {
		Enable bool `yaml:"enable"`
	} `yaml:"trace-collect"`
	ProfileCollect struct {
		Enable      bool `yaml:"enable"`
		RepeatTimes int  `yaml:"repeat-times"`
	}

	Cases []struct {
		Title     string `yaml:"title"`
		Suite     string `yaml:"suite"`
		Directory string `yaml:"directory"`
		Command   string `yaml:"command"`
		Args      []struct {
			Param1 string `yaml:"param1"`
			Param2 string `yaml:"param2"`
		} `yaml:"args"`
	} `yaml:"cases"`
}

func (c *SimSetting) load(file string) {
	file = filepath.Join(projectRoot, file)
	bytes, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		log.Panic(err)
	}
}
