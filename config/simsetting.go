package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SimSetting struct {
	DeviceID int `yaml:"device-id"`

	RepeatTimes    int  `yaml:"repeat-times"`
	UpdateToServer bool `yaml:"update-to-server"`

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
		log.Panic(err)
	}
	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		log.Panic(err)
	}
}
