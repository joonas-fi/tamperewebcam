package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Bucket(bucket string, region string) (*BucketContext, error) {
	awsSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(
		awsSession,
		aws.NewConfig().WithRegion(region))

	return &BucketContext{
		Name: &bucket,
		S3:   s3Client,
	}, nil
}

type BucketContext struct {
	Name *string
	S3   *s3.S3
}
