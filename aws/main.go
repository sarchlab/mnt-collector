package aws

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/sarchlab/mnt-collector/config"
)

const traceDir = "./tmp/mnt-traces/"

var mntClient *s3.Client
var mntBucket *string

func Connect() {
	mntClient = s3.New(s3.Options{
		Region: config.SC.AWS.Region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			config.SC.AWS.AccessKeyID, config.SC.AWS.SecretAccessKey, "")),
	})
	mntBucket = aws.String(config.SC.AWS.Bucket)

	_, err := ListObjects("")
	if err != nil {
		log.WithError(err).Panic("Failed to connect to AWS")
	}

	err = os.MkdirAll(traceDir, 0755)
	if err != nil {
		log.WithField("traceDir", traceDir).WithError(err).
			Fatal("Could not create local directory for traces")
	}
}
