package main

import (
	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/collector"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	collector.LoadDevice(config.C.DeviceID)
	log.Info("Device loaded.")

	mntbackend.Init()
	log.Info("MNT backend connected.")

	aws.Init()
	log.Info("AWS connected.")

	// collector.Start()
}
