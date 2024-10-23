package collector

import (
	log "github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
)

type profileRawData struct {
	KernelName string `db:"kernelName"`
	StartTime  int64  `db:"start"`
	EndTime    int64  `db:"end"`
}

func openDB(profileFile string) *sqlx.DB {
	db, err := sqlx.Open("sqlite3", profileFile)
	if err != nil {
		log.WithError(err).Error("Failed to open profile file")
	}
	return db
}

func getKernelActivities(db *sqlx.DB) ([]profileRawData, error) {
	query := `SELECT k.start, k.end, s.value AS kernelName
		FROM CUPTI_ACTIVITY_KIND_KERNEL k
		JOIN StringIds s ON k.demangledName = s.id;`

	var results []profileRawData
	err := db.Select(&results, query)
	if err != nil {
		log.WithError(err).Error("Failed to get kernel activities")
		return nil, err
	}

	return results, nil
}
