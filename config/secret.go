package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SecretConfig struct {
	MNT struct {
		Host  string `yaml:"host"`
		Port  int    `yaml:"port"`
		Base  string `yaml:"base"`
		Token string `yaml:"token"`
	} `yaml:"mnt-backend"`

	AWS struct {
		Region          string `yaml:"region"`
		Bucket          string `yaml:"bucket"`
		AccessKeyID     string `yaml:"access-key-id"`
		SecretAccessKey string `yaml:"secret-access-key"`
	} `yaml:"s3"`
}

func (c *SecretConfig) load(file string) {
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
