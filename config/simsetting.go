package config

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
)

type Case struct {
	Title     string    `yaml:"title"`
	Suite     string    `yaml:"suite"`
	Directory string    `yaml:"directory"`
	Command   string    `yaml:"command"`
	Args      []CaseArg `yaml:"args"`
}

type CaseArg struct {
	Size       int32 `yaml:"size" json:"size"`
	VectorN    int32 `yaml:"vectorN" json:"vectorN"`
	ElementN   int32 `yaml:"elementN" json:"elementN"`
	Log2Data   int32 `yaml:"log2data" json:"log2data"`
	Log2Kernel int32 `yaml:"log2kernel" json:"log2kernel"`
	DimX       int32 `yaml:"dimX" json:"dimX"`
	DimY       int32 `yaml:"dimY" json:"dimY"`
	Sizemult   int32 `yaml:"sizemult" json:"sizemult"`
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
