package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rancher/rke/metadata"
)

const (
	defaultURL = "https://raw.githubusercontent.com/vardhaman22/rancher-kontainer-driver-metadata/update-default-version/data/data.json"
	dataFile   = "data/data.json"
)

// Codegen fetch data.json from https://releases.rancher.com/kontainer-driver-metadata/release-v2.7/data.json and generates bindata
func main() {
	u := os.Getenv(metadata.RancherMetadataURLEnv)
	if u == "" {
		u = defaultURL
	}
	data, err := http.Get(u)
	if err != nil {
		panic(fmt.Errorf("failed to fetch data.json from kontainer-driver-metadata repository"))
	}
	defer data.Body.Close()

	b, err := io.ReadAll(data.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Writing data")
	if err := os.WriteFile(dataFile, b, 0755); err != nil {
		return
	}
}
