package main

import (
	"fmt"

	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
)

func main() {
	config.C.Load("etc/simsetting.yaml")
	config.SC.Load("etc/secret.yaml")
	fmt.Println("Config loaded.")

	mntbackend.Init()
	fmt.Println("MNT backend connected.")

	aws.Init()
	fmt.Println("AWS connected.")

	
}
