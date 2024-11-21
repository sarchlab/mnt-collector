package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/transport/http"
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

func CheckObjectExist(objectKey string) (bool, error) {
	_, err := mntClient.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: mntBucket,
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var respErr *http.ResponseError
		if errors.As(err, &respErr) && respErr.Response != nil && respErr.Response.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
