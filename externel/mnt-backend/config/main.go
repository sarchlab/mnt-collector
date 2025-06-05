package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DB struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"mongodb"`

	Server struct {
		Port string `yaml:"port"`
		Base string `yaml:"base"`
		Auth struct {
			ReadToken  string `yaml:"read"`
			WriteToken string `yaml:"write"`
		} `yaml:"authentication"`
	} `yaml:"server"`
}

var C *Config

func Init(file string) {
	var root string
	dir, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			root = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Panic("config: could not find project root")
		}
		dir = parent
	}

	C = &Config{}
	err = C.load(filepath.Join(root, file))

	if err != nil {
		log.Panic(err)
	}
}

func (c *Config) load(file string) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		return err
	}
	return nil
}
