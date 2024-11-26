package aws

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

func getObject(object string, filepath string) error {
	result, err := mntClient.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: mntBucket,
		Key:    aws.String(object),
	})
	if err != nil {
		log.WithError(err).Error("Could not get object")
		return err
	}
	file, err := os.Create(filepath)
	if err != nil {
		log.WithError(err).Error("Could not create file")
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, result.Body)
	if err != nil {
		log.WithError(err).Error("Could not copy object to file")
		return err
	}
	return nil
}

func listObjects(prefix string) ([]string, error) {
	result, err := mntClient.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: mntBucket,
		Prefix: aws.String(prefix),
	})
	if err != nil {
		log.WithError(err).Error("Could not list objects")
		return nil, err
	}
	var objects []string
	for _, obj := range result.Contents {
		objects = append(objects, *obj.Key)
	}
	return objects, nil
}

func SyncDirToLocal(prefix string) (string, error) {
	objects, err := listObjects(prefix)
	if err != nil {
		log.WithField("prefix", prefix).WithError(err).Error("Could not list objects")
		return "", err
	}

	localDir := filepath.Join(traceDir, prefix)
	err = os.MkdirAll(localDir, 0755)
	if err != nil {
		log.WithError(err).Error("Could not create local directory")
		return "", err
	}

	for _, obj := range objects {
		objectBase, err := filepath.Rel(prefix, obj)
		if err != nil {
			log.WithFields(log.Fields{
				"object": obj,
				"prefix": prefix,
			}).WithError(err).Error("cannot get relative path")
			return "", err
		}
		localFile := filepath.Join(localDir, objectBase)
		log.WithFields(log.Fields{
			"object_path": obj,
			"local_path":  localFile,
		}).Debug("Checking object")
		if fileExist(localFile) {
			continue
		}
		log.WithField("object", obj).Info("Downloading object")
		err = getObject(obj, localFile)
		if err != nil {
			log.WithField("object", obj).WithError(err).Error("Could not get object")
			return "", err
		}
	}

	return localDir, nil
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}
