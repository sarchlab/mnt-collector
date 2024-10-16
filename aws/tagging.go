package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func GetObjectTagging(object string) ([]types.Tag, error) {
	output, err := mntClient.GetObjectTagging(context.TODO(), &s3.GetObjectTaggingInput{
		Bucket: mntBucket,
		Key:    aws.String(object),
	})
	if err != nil {
		return nil, err
	}

	return output.TagSet, nil
}

func SetObjectTagging(object string, tags []types.Tag) error {
	_, err := mntClient.PutObjectTagging(context.TODO(), &s3.PutObjectTaggingInput{
		Bucket:  mntBucket,
		Key:     aws.String(object),
		Tagging: &types.Tagging{TagSet: tags},
	})
	if err != nil {
		return err
	}

	return nil
}
