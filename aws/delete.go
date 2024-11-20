package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	log "github.com/sirupsen/logrus"
)

func DeleteObjects(objects []types.ObjectIdentifier) error {
	_, err := mntClient.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: mntBucket,
		Delete: &types.Delete{
			Objects: objects,
		},
	})

	return err
}

func DeleteObjectDirectory(objectDir string) error {
	log.WithField("objectDir", objectDir).Debug("Deleting object directory")
	objects, err := ListObjects(objectDir)
	if err != nil {
		return err
	}

	objIdentifiers := []types.ObjectIdentifier{}
	for _, object := range objects {
		objIdentifiers = append(objIdentifiers, types.ObjectIdentifier{Key: &object})
	}

	return DeleteObjects(objIdentifiers)
}
