package collector

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sarchlab/mnt-backend/model"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
)

const profilesDir = "./tmp/mnt-collector/profiles/"

type ProfileData struct {
	AvgNanoSec   float64
	Frequency    uint32
	MaxFrequency uint32
}

func RunProfileCollection() {
	caseSettings := generateCaseSettings(config.C.Cases)
	repeatTimes := config.C.RepeatTimes

	err := os.MkdirAll(profilesDir, 0755)
	if err != nil {
		log.WithError(err).Error("Failed to create profile directory")
	}

	for _, c := range caseSettings {
		log.WithFields(log.Fields{
			"Title":   c.Title,
			"Suite":   c.Suite,
			"Command": c.Command,
			"Params":  c.ParamStr,
		}).Info("Start profile collection")

		var profileFiles []string
		for i := 0; i < int(repeatTimes); i++ {
			log.WithField("Repeat", i).Info("Start profiling")

			profileFile, err := profile(c)
			if err != nil {
				log.WithError(err).Error("Failed to profile")
				break
			}

			log.WithField("ProfileFile", profileFile).Info("Profiled")
			profileFiles = append(profileFiles, fmt.Sprintf("%s.sqlite", profileFile))
		}

		if len(profileFiles) == int(repeatTimes) {
			log.Info("Start processing profile")
			data := getProfileData(profileFiles)

			if config.C.UploadToServer {
				log.Info("Start uploading to server")
				uploadProfileToDB(c, data, repeatTimes)
			} else {
				log.Info("Skip uploading to server")
			}
		} else {
			log.Error("Profile not completed, skip processing")
		}
	}
}

func profile(c CaseSetting) (string, error) {
	file, err := os.CreateTemp(profilesDir, "profile-*.nsys")
	if err != nil {
		log.WithError(err).Error("Failed to create profile file")
		return "", err
	}

	getCmd := func() *exec.Cmd {
		param := strings.Split(c.ParamStr, " ")
		args := append([]string{"profile", "--stats=true", "--output=" + file.Name(), c.Command}, param...)
		// cmd := exec.Command("nsys", "profile", "--stats=true", "--output="+file.Name(), c.Command, param...)
		cmd := exec.Command("nsys", args...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("CUDA_VISIBLE_DEVICES=%d", config.C.DeviceID))
		return cmd
	}

	log.Info("Start profiling")
	for err := runGPUCmdWithTimer(getCmd()); err != nil; {
		if err == ErrorInterrupted {
			log.Warn("Interrupted, retry profiling")
			waitTillDeviceIdle()
			err = runGPUCmdWithTimer(getCmd())
		} else {
			log.WithError(err).Error("Failed to run command")
			return "", err
		}
	}

	return file.Name(), nil
}

func getProfileData(profileFiles []string) ProfileData {
	var sumTime int64

	for _, file := range profileFiles {
		profileFile := openDB(file)
		activities, err := getKernelActivities(profileFile)
		if err != nil {
			log.WithError(err).Error("Failed to get kernel activities")
		}

		log.WithField("profileFile", file).Debug("Processing profile")
		for _, kernel := range activities {
			// log.WithFields(log.Fields{
			// 	"KernelName": kernel.KernelName,
			// 	"StartTime":  kernel.StartTime,
			// 	"EndTime":    kernel.EndTime,
			// }).Debug("Kernel activity")
			sumTime += kernel.EndTime - kernel.StartTime
		}
		// nowAvgNanoSec := float64(sumTime) / float64(i+1)
		// log.WithField("AvgNanoSec", nowAvgNanoSec).Debug("Current average time")
		profileFile.Close()
	}

	repeatTimes := len(profileFiles)
	avgNanoSec := float64(sumTime) / float64(repeatTimes)
	//	avgCycles := float64(sumTime) / float64(repeatTimes) / (1e9 * float64(config.Frequency()))

	data := ProfileData{
		AvgNanoSec:   avgNanoSec,
		Frequency:    config.Frequency(),
		MaxFrequency: config.MaxFrequency(),
		// AvgCycles: avgCycles,
	}
	log.WithFields(log.Fields{
		"avgNanoSec": avgNanoSec,
	}).Info("Profile data")

	return data
}

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

func uploadProfileToDB(c CaseSetting, data ProfileData, repeatTimes int32) {
	req := model.DBProf{
		CaseKey: model.CaseKey{
			EnvID:     mntbackend.EnvID(),
			Suite:     c.Suite,
			Benchmark: c.Title,
			Param:     c.param,
		},
		RepeatTimes:  repeatTimes,
		AvgNanoSec:   data.AvgNanoSec,
		Frequency:    data.Frequency,
		MaxFrequency: data.MaxFrequency,
	}
	profileID, err := mntbackend.UpdOrUplProfile(req)
	if err != nil {
		log.WithFields(log.Fields{
			"Case": c,
			"Data": data,
		}).WithError(err).Error("Failed to upload profile")
	} else {
		log.WithField("ProfileID", profileID).Info("Profile uploaded")
	}
}
