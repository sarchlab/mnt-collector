package config

import (
	log "github.com/sirupsen/logrus"
)

var C *CollectionSettings
var SC *SecretConfig

func init() {
	prepareBasicEnvirons()

	C = &CollectionSettings{}
	SC = &SecretConfig{}
}

func LoadConfig(collectFile string, secretFile string) {
	log.WithField("secretFile", secretFile).Info("Loading secret tokens")
	SC.load(collectFile)

	log.WithField("collectFile", collectFile).Info("Loading collection settings")
	C.load(collectFile)
}
