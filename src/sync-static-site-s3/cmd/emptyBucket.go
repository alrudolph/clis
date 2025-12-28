package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func EmptyBucket(bucketName, prefix string, client *s3.Client, ctx context.Context) error {
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: &prefix,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return err
		}

		var objects []types.ObjectIdentifier

		for _, obj := range page.Contents {
			fmt.Printf("> removing object %s\n", *obj.Key)
			objects = append(objects, types.ObjectIdentifier{Key: obj.Key})
		}

		if len(objects) == 0 {
			continue
		}

		_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: objects,
			},
		})

		if err != nil {
			return nil
		}
	}

	return nil
}
