package main

import (
	"os"
	"testing"

	"github.com/rancher/rke/metadata"
)

func TestGetURL(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		tag         string
		expectedURL string
	}{
		{
			name:        "No Metadata URL and TAG is release version",
			envVar:      "",
			tag:         "v1.0.0",
			expectedURL: defaultReleaseURL,
		},
		{
			name:        "No Metadata URL and TAG is pre-release version",
			envVar:      "",
			tag:         "v1.0.0-alpha",
			expectedURL: defaultDevURL,
		},
		{
			name:        "Metadata URL set",
			envVar:      "https://example.com",
			tag:         "v1.0.0",
			expectedURL: "https://example.com",
		},
		{
			name:        "Invalid TAG",
			envVar:      "",
			tag:         "invalid-tag",
			expectedURL: defaultDevURL,
		},
		{
			name:        "No TAG",
			envVar:      "",
			tag:         "",
			expectedURL: defaultDevURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variables
			os.Setenv(metadata.RancherMetadataURLEnv, tt.envVar)
			os.Setenv("TAG", tt.tag)
			defer func() {
				os.Unsetenv(metadata.RancherMetadataURLEnv)
				os.Unsetenv("TAG")
			}()

			result := getURL()

			if result != tt.expectedURL {
				t.Errorf("expected %s, got %s", tt.expectedURL, result)
			}
		})
	}
}
