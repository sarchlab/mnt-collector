package collector

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sarchlab/mnt-backend/model"
	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	"go.mongodb.org/mongo-driver/bson/primitive"

	log "github.com/sirupsen/logrus"
)

type SimData struct {
	PredictCycle int
}

func RunSimulationCollection() {
	traceIDs := config.C.TraceIDs

	for _, traceID := range traceIDs {
		dbTrace, err := mntbackend.GetTraceByID(traceID)
		if err != nil {
			log.WithError(err).WithField("trace_id", traceID.Hex()).Fatal("Could not get trace")
		}

		traceDir, err := aws.SyncDirToLocal(dbTrace.S3Path)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"trace_id": traceID.Hex(),
				"s3_path":  dbTrace.S3Path,
			}).Fatal("Could not sync trace")
		}

		res, err := simulate(traceDir)
		if err != nil {
			log.WithError(err).WithField("trace_id", traceID.Hex()).Fatal("Could not simulate")
		}
		log.WithField("result", res).Debug("Simulation finished")

		if config.C.UploadToServer {
			log.Info("uploading to server")
			uploadSimToDB(traceID, res)
		} else {
			log.Info("skip uploading to server")
			fmt.Println(res.PredictCycle)
		}
	}
}

func simulate(traceDir string) (SimData, error) {

	// Check if traceDir/traces exists and is a directory
	tracesPath := filepath.Join(traceDir, "traces")
	info, err := os.Stat(tracesPath)
	if err == nil && info.IsDir() {
		traceDir = tracesPath
	}
	cmdStr := fmt.Sprintf("-trace-dir %s", traceDir)
	cmdArgs := strings.Split(cmdStr, " ")
	cmd := exec.Command(config.C.Experiment.Runfile, cmdArgs...)

	var outBuff bytes.Buffer
	cmd.Stdout = &outBuff

	err = runNormalCmdWithTimer(cmd)
	if err != nil {
		log.Panic("Failed to run simulation")
	}

	output := strings.TrimSpace(outBuff.String())
	// predictCycle, err := strconv.ParseInt(output, 10, 32)
	predictCycleFloat, err := strconv.ParseFloat(output, 64)

	data := SimData{
		PredictCycle: int(predictCycleFloat),
	}

	return data, err
}

func uploadSimToDB(traceID primitive.ObjectID, res SimData) {
	req := model.DBSim{
		SimKey: model.SimKey{
			ExpID:   mntbackend.ExpID(),
			TraceID: traceID,
		},
		PredictCycle: float64(res.PredictCycle),
	}
	simID, err := mntbackend.UpdOrUplSim(req)
	if err != nil {
		log.WithFields(log.Fields{
			"trace_id": traceID.Hex(),
			"result":   res,
		}).WithError(err).Error("Could not upload simulation")
	} else {
		log.WithField("simID", simID.Hex()).Info("Simulation uploaded")
	}
}
