package cmd

import "testing"

func TestGetObjectKeys(t *testing.T) {
	tests := []struct {
		fileName          string
		expected          string
		expectedExtension string
	}{
		{"index.html", "index.html", "text/html; charset=utf-8"},
		{"error.html", "error.html", "text/html; charset=utf-8"},
		{"file.html", "file", "text/html; charset=utf-8"},
		{"styles.css", "styles.css", "text/css; charset=utf-8"},
		{"data.json", "data.json", "application/json"},
		{"script.js", "script.js", "text/javascript; charset=utf-8"},
	}

	for _, test := range tests {
		objectKey, objectExtension := getObjectKeyType(test.fileName)
		if objectKey != test.expected || objectExtension != test.expectedExtension {
			t.Errorf("expected %s (%s), got %s (%s)", test.expected, test.expectedExtension, objectKey, objectExtension)
		}
	}
}
