package config

import log "github.com/sirupsen/logrus"

var (
	nsysVersion string
)

func NsysVersion() string {
	if nsysVersion == "" {
		log.Panic("nsysVersion is not initialized")
	}
	return nsysVersion
}

func PrepareProfilesEnvirons() {
	nsysVersion = getNsysVersion()
}
