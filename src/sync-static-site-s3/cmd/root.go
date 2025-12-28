package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

type Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Profile         string
	Role            string
	Bucket          string
	Prefix          string
	Directory       string
	CfInvalidate    bool
}

type SavedConfig struct {
	UserDirectory   string `json:"userDirectory"`
	Name            string `json:"name"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Profile         string `json:"profile"`
	Role            string `json:"role"`
	Bucket          string `json:"bucket"`
	Directory       string `json:"directory"`
}

type SavedConfigFile struct {
	Profiles []SavedConfig `json:"profiles"`
}

func LoadConfigFromFile(configName string) (*Config, error) {
	profiles, err := LoadConfigOptions()

	if err != nil {
		return nil, err
	}

	var foundProfile *SavedConfig = nil

	for _, profile := range profiles {
		if profile.Name == configName {
			foundProfile = &profile
			break
		}
	}

	if foundProfile == nil {
		return nil, fmt.Errorf("config with name %s not found", configName)
	}

	return &Config{
		Region:          foundProfile.Region,
		AccessKeyID:     foundProfile.AccessKeyID,
		SecretAccessKey: foundProfile.SecretAccessKey,
		Profile:         foundProfile.Profile,
		Role:            foundProfile.Role,
		Bucket:          foundProfile.Bucket,
		Directory:       foundProfile.Directory,
	}, nil
}

func NewConfig(cmd *cobra.Command, args []string) (*Config, error) {
	configName, loadFromConfigErr := cmd.Flags().GetString("config")

	if loadFromConfigErr == nil && configName != "" {
		config, err := LoadConfigFromFile(configName)

		if err != nil {
			log.Fatal(err)
		}

		return config, nil
	}

	directory, _ := cmd.Flags().GetString("directory")
	bucket, _ := cmd.Flags().GetString("bucket")
	prefix, _ := cmd.Flags().GetString("prefix")

	if bucket == "" {
		return nil, errors.New("bucket is required")
	}

	region, _ := cmd.Flags().GetString("region")

	// Credentials:
	profile, _ := cmd.Flags().GetString("profile")
	accessKeyId, _ := cmd.Flags().GetString("access-key-id")
	secretAccesKey, _ := cmd.Flags().GetString("secret-access-key")
	role, _ := cmd.Flags().GetString("role")

	cfInvalidate, _ := cmd.Flags().GetBool("cf-invalidate")

	return &Config{
		Region:          region,
		AccessKeyID:     accessKeyId,
		SecretAccessKey: secretAccesKey,
		Profile:         profile,
		Role:            role,
		Bucket:          bucket,
		Directory:       directory,
		Prefix:          prefix,
		CfInvalidate:    cfInvalidate,
	}, nil
}

var RootCmd = &cobra.Command{
	Use:   "sync-static-site-s3",
	Short: "Upload a directory containing files for a static site to a S3 Bucket",
	Long: `This CLI uploads the files in a directory to an S3 bucket but makes important
considerations for html files. If the file is an html file, the .html extension is removed
except for index.html and error.html, since we no longer have the html extension we also need
to set the file's Content-Type. This is necessary to allow the files to be accessed without
the .html extension, for example, domain.com/file instead of domain.com/file.html.

Example Usage:
	go run . --directory /path/to/static/site --bucket s3-bucket-name
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		if len(args) > 0 {
			fmt.Println("Additional supplied args will be ignored")
		}

		userInput, err := NewConfig(cmd, args)

		if err != nil {
			log.Fatal(err)
		}

		_, awsConfig, err := GetAWSConfig(
			userInput.AccessKeyID,
			userInput.SecretAccessKey,
			userInput.Profile,
			userInput.Region,
			userInput.Role,
			ctx,
		)

		if err != nil {
			log.Fatal(err)
		}

		client := s3.NewFromConfig(awsConfig)

		err = EmptyBucket(userInput.Bucket, userInput.Prefix, client, ctx)

		if err != nil {
			fmt.Println("Failed to clear bucket, aborting upload")
			log.Fatal(err)
		}

		if userInput.Directory == "" {
			return
		}

		// TODO: can i batch this?
		err = UploadDirectory(
			userInput.Directory,
			userInput.Bucket,
			userInput.Prefix,
			client,
			ctx,
		)
		if err != nil {
			log.Fatal(err)
		}

		if !userInput.CfInvalidate {
			return
		}

		fmt.Println("Creating CloudFront invalidation...")

		cloudFrontClient := cloudfront.NewFromConfig(awsConfig)

		_, err = InvalidateCache(userInput.Bucket, cloudFrontClient, ctx)

		if err != nil {
			log.Fatal(err)
		}
	},
}

func UploadDirectory(directory, bucket, prefix string, client *s3.Client, ctx context.Context) error {
	return filepath.Walk(
		directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			return UploadFile(directory, path, bucket, prefix, client, ctx)
		},
	)
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.Flags().StringP("config", "c", "", "Config Profile to use. See config subcommand to list options.")
	RootCmd.Flags().StringP("directory", "d", "", "Path to the static site directory")
	_ = RootCmd.MarkFlagDirname("directory")
	RootCmd.Flags().StringP("bucket", "b", "", "S3 bucket name")
	RootCmd.Flags().StringP("prefix", "x", "", "S3 bucket path prefix")
	RootCmd.Flags().StringP("region", "r", "us-east-1", "S3 bucket region")
	RootCmd.Flags().String("access-key-id", "", "AWS Access Key ID")
	RootCmd.Flags().String("secret-access-key", "", "AWS Secret Access Key")
	RootCmd.Flags().StringP("profile", "p", "", "AWS Profile name")
	RootCmd.Flags().StringP("role", "", "", "Role to switch into")
	RootCmd.Flags().BoolP("cf-invalidate", "", false, "Wether to create a CloudFront invalidation")
}
