package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func GetAWSConfig(accessKeyId, secretAccessKey, profile, region, roleName string, ctx context.Context) (string, aws.Config, error) {
	profile, config, err := getAWSConfig(accessKeyId, secretAccessKey, profile, region, ctx)

	if err != nil {
		return "", aws.Config{}, err
	}

	if roleName == "" {
		return profile, config, nil
	}

	// handle role switching:
	stsClient := sts.NewFromConfig(config)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleName)
	config.Credentials = aws.NewCredentialsCache(provider)

	return profile, config, nil
}

func getAWSConfig(accessKeyId, secretAccessKey, profile, region string, ctx context.Context) (string, aws.Config, error) {
	// load from profile OR use access key/secret access key (cannot supply both)
	// otherwise, try to use the $AWS_PROFILE profile

	if profile != "" {
		if accessKeyId != "" || secretAccessKey != "" {
			return "", aws.Config{}, errors.New("cannot provide both profile and access key id/secret access key")
		}
		fmt.Printf("Using profile %s\n", profile)
		c, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))

		if err == nil && c.Region == "" {
			c.Region = region
		}

		return profile, c, err
	}

	if accessKeyId != "" && secretAccessKey != "" {
		fmt.Println("Using access keys")
		return "", aws.Config{
			Region: region,
			Credentials: credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     accessKeyId,
					SecretAccessKey: secretAccessKey,
				},
			},
		}, nil
	}

	profile = os.Getenv("AWS_PROFILE")

	if profile != "" {
		fmt.Printf("Using default profile %s\n", profile)
		c, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))

		if err == nil && c.Region == "" {
			c.Region = region
		}

		return profile, c, err
	}

	accessKeyId, present := os.LookupEnv("AWS_ACCESS_KEY_ID")

	if !present {
		return "", aws.Config{}, errors.New("no access key id provided")
	}

	secretAccessKey, present = os.LookupEnv("AWS_SECRET_ACCESS_KEY")

	if !present {
		return "", aws.Config{}, errors.New("no secret access key provided")
	}

	fmt.Println("Using access keys from environment variables")
	return "", aws.Config{
		Region: region,
		Credentials: credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKeyId,
				SecretAccessKey: secretAccessKey,
			},
		},
	}, nil
}
func LoadConfigOptions() ([]SavedConfig, error) {
	usr, err := user.Current()

	if err != nil {
		return nil, err
	}

	homeDir := usr.HomeDir

	configFile, err := os.Open(filepath.Join(homeDir, "sync-s3", "config.json"))

	if err != nil {
		return nil, errors.New("no config profiles found, create one using setup subcommand")
	}

	savedConfig := SavedConfigFile{}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&savedConfig); err != nil {
		return nil, err
	}

	cwd, err := filepath.Abs(".")

	if err != nil {
		return nil, err
	}

	profiles := []SavedConfig{}

	for _, profile := range savedConfig.Profiles {
		if profile.UserDirectory != cwd {
			continue
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}
