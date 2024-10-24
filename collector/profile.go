package collector

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
)

const profilesDir = "/tmp/mnt-collector/profiles/"

type ProfileData struct {
	AvgNanoSec   float64
	Frequency    uint32
	MaxFrequency uint32
}

func runProfileCollect(cases []Case) {
	err := os.MkdirAll(profilesDir, 0755)
	if err != nil {
		log.WithError(err).Error("Failed to create profile directory")
	}

	for _, c := range cases {
		log.WithFields(log.Fields{
			"Title":       c.Title,
			"Suite":       c.Suite,
			"Command":     c.Command,
			"Params":      c.ParamStr,
			"RepeatTimes": c.RepeatTimes,
		}).Info("Start profile collection")

		var profileFiles []string
		for i := 0; i < int(c.RepeatTimes); i++ {
			log.WithField("Repeat", i).Info("Start profiling")

			profileFile, err := profile(c)
			if err != nil {
				log.WithError(err).Error("Failed to profile")
				break
			}

			log.WithField("ProfileFile", profileFile).Info("Profiled")
			profileFiles = append(profileFiles, fmt.Sprintf("%s.sqlite", profileFile))
		}

		if len(profileFiles) == int(c.RepeatTimes) {
			log.Info("Start processing profile")
			data := getProfileData(profileFiles)

			if config.C.UploadToServer {
				log.Info("Start uploading to server")
				uploadProfileToDB(c, data)
			}

		} else {
			log.Error("Profile not completed, skip processing")
		}
	}
}

func profile(c Case) (string, error) {
	file, err := os.CreateTemp(profilesDir, "profile-*.nsys")
	if err != nil {
		log.WithError(err).Error("Failed to create profile file")
		return "", err
	}

	cmd := exec.Command("nsys", "profile", "--stats=true", "--output="+file.Name(), c.Command, c.ParamStr)
	cmd.Env = append(os.Environ(), fmt.Sprintf("CUDA_VISIBLE_DEVICES=%d", config.C.DeviceID))

	log.Info("Start profiling")
	err = runCommandWithTimer(cmd)
	if err != nil {
		log.Error("Failed to run command")
		return "", err
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

		for _, kernel := range activities {
			sumTime += kernel.EndTime - kernel.StartTime
		}
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

func uploadProfileToDB(c Case, data ProfileData) {
	req := mntbackend.ProfileRequest{
		EnvID:       mntbackend.EnvID(),
		Suite:       c.Suite,
		Benchmark:   c.Title,
		Param:       c.param,
		RepeatTimes: c.RepeatTimes,

		AvgNanoSec:   data.AvgNanoSec,
		Frequency:    data.Frequency,
		MaxFrequency: data.MaxFrequency,
	}
	profileID, err := mntbackend.UploadProfile(req)
	if err != nil {
		log.WithFields(log.Fields{
			"Case": c,
			"Data": data,
		}).WithError(err).Error("Failed to upload profile")
	} else {
		log.WithField("ProfileID", profileID).Info("Profile uploaded")
	}
}
