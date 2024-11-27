package mntbackend

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/sarchlab/mnt-collector/config"
)

var (
	ErrorStatusNotOK = errors.New("response status not OK")
	ErrorNilData     = errors.New("response data is nil")
	ErrorNotHealthy  = errors.New("mnt backend health check failed")
	ObjectNotFound   = errors.New("not found")
)

var URLBase string

func Connect() {
	c := config.SC.MNT
	URLBase = fmt.Sprintf("http://%s:%d%s", c.Host, c.Port, c.Base)

	err := checkHealth()
	if err != nil {
		log.WithError(err).Panic("Failed to connect to MNT backend")
	}

}
