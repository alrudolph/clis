package cmd

import (
	"log"
	"testing"
)

func TestGetAWSConfig(t *testing.T) {
	tests := []struct {
		accessKeyId     string
		secretAccessKey string
		profile         string
		region          string
		expectedError   bool
	}{
		{"", "", "", "", false},
	}

	for _, test := range tests {
		_, creds, err := GetAWSConfig(test.accessKeyId, test.secretAccessKey, test.profile, test.region, "", nil)

		if test.expectedError && err == nil {
			log.Fatalf("Test did not fail")
		}

		if !test.expectedError && err != nil {
			log.Fatalf("Test failed: %s", err)
		}

		// TODO: test profile or keys set
		_ = creds
	}
}
