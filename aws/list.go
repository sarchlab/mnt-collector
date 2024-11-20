package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ListObjects(objectPrefix string) ([]string, error) {
	output, err := mntClient.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: mntBucket,
		Prefix: aws.String(objectPrefix),
	})
	if err != nil {
		return nil, err
	}

	var objects []string
	for _, object := range output.Contents {
		objects = append(objects, *object.Key)
	}
	return objects, nil
}
