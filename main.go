package main

import (
	"fmt"

	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/collector"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
)

func main() {
	config.LoadSimSetting("etc/simsetting.yaml")
	config.LoadSecretConfig("etc/secret.yaml")
	fmt.Println("Config loaded.")

	collector.LoadDevice(config.C.DeviceID)
	fmt.Println("Environment checked.")

	mntbackend.Init()
	fmt.Println("MNT backend connected.")

	aws.Init()
	fmt.Println("AWS connected.")
}
