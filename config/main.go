package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GPU         string `yaml:"gpu"`
	CudaVersion string `yaml:"cuda-version"`
	Machine     string `yaml:"machine"`

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

var projectRoot string
var C *Config
var SC *SecretConfig

func init() {
	dir, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			projectRoot = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Panic("config: could not find project root")
		}
		dir = parent
	}

	C = &Config{}
	SC = &SecretConfig{}
}

func (c *Config) Load(file string) {
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

func (c *SecretConfig) Load(file string) {
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
