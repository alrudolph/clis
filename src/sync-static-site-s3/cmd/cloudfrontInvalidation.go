package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

func InvalidateCache(bucketName string, client *cloudfront.Client, ctx context.Context) (*cloudfront.CreateInvalidationOutput, error) {
	// get distribution id
	distributionID, err := getDistributionID(bucketName, client, ctx)

	if err != nil {
		return nil, err
	}

	// create invalidation
	invalidation := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(distributionID),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			Paths: &types.Paths{
				Quantity: aws.Int32(1),
				Items:    []string{"/*"},
			},
		},
	}

	return client.CreateInvalidation(ctx, invalidation)
}

func getDistributionID(bucketName string, client *cloudfront.Client, ctx context.Context) (string, error) {
	// get distribution id
	distributionList, err := client.ListDistributions(ctx, &cloudfront.ListDistributionsInput{})

	if err != nil {
		return "", err
	}

	expectedDomainName := fmt.Sprintf("%s.s3.%s.amazonaws.com", bucketName, client.Options().Region)

	for _, distribution := range distributionList.DistributionList.Items {
		if *distribution.Origins.Items[0].DomainName == expectedDomainName {
			return *distribution.Id, nil
		}
	}

	return "", fmt.Errorf("distribution for bucket %s not found", bucketName)
}
