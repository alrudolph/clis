package cmd

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadFile(baseDirectory, path, bucketName, prefix string, client *s3.Client, ctx context.Context) error {
	fileName, err := filepath.Rel(baseDirectory, path)

	if err != nil {
		return err
	}

	keyName, mimeType := getObjectKeyType(fileName)

	if prefix != "" {
		keyName = filepath.Join(prefix, keyName)
	}

	fmt.Printf("> uploading %s - %s\n", keyName, mimeType)

	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	obj := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(keyName),
		Body:        file,
		ContentType: aws.String(mimeType),
	}

	_, err = client.PutObject(ctx, obj)

	return err
}

func getObjectKeyType(fileName string) (outputFileName, mimeType string) {
	// html files should not have the .html extension as that will
	// require use to access domain.com/file.html instead of domain.com/file
	// we make an exception for "index.html" and "error.html"
	mimeType = mime.TypeByExtension(filepath.Ext(fileName))
	outputFileName = fileName

	if fileName == "index.html" || fileName == "error.html" {
		return
	}

	if filepath.Ext(fileName) == ".html" {
		outputFileName = fileName[:len(fileName)-5]
	}

	return
}
