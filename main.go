package main

import (
	"io"
	"os"

	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/collector"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
)

func main() {
	setLog()

	config.LoadDevice(config.C.DeviceID)
	log.Info("Device loaded.")

	mntbackend.Init()
	log.Info("MNT backend connected.")

	aws.Init()
	log.Info("AWS connected.")

	log.Info("Start collecting data.")
	collector.Run()
}

func setLog() {
	file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	multiWriter := io.MultiWriter(file, os.Stdout)

	log.SetOutput(multiWriter)
	log.SetLevel(log.DebugLevel)
}
