package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/rancher/rke/metadata"
)

const (
	defaultDevURL     = "https://raw.githubusercontent.com/chiukapoor/kontainer-driver-metadata/refs/heads/v2.9-oct-24/data/data.json"
	defaultReleaseURL = "https://releases.rancher.com/kontainer-driver-metadata/release-v2.9/data.json"
	dataFile          = "data/data.json"
)

// Codegen fetch data.json from defaultURL or defaultReleaseURL and generates bindata
func main() {
	u := getURL()
	fmt.Printf("Reading data from %s \n", u)
	data, err := http.Get(u)
	if err != nil {
		panic(fmt.Errorf("failed to fetch data.json from kontainer-driver-metadata repository"))
	}
	defer data.Body.Close()

	b, err := io.ReadAll(data.Body)
	if err != nil {
		panic(err)
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(b, &jsonData)
	if err != nil {
		panic(fmt.Errorf("failed to parse JSON: %v", err))
	}
	// rke doesn't need info about rke2 and k3s versions
	delete(jsonData, "rke2")
	delete(jsonData, "k3s")

	b, err = json.Marshal(jsonData)
	if err != nil {
		panic(fmt.Errorf("failed to marshal json data: %v", err))
	}

	fmt.Println("Writing data")
	if err := os.WriteFile(dataFile, b, 0755); err != nil {
		return
	}
}

func getURL() string {
	u := os.Getenv(metadata.RancherMetadataURLEnv)
	if u == "" {
		u = defaultDevURL
		tag := os.Getenv("TAG")
		if strings.HasPrefix(tag, "v") {
			tag = tag[1:]
		}
		if v, err := semver.NewVersion(tag); err == nil {
			if v.PreRelease == "" && v.String() != "" {
				u = defaultReleaseURL
			}
		}
	}
	return u
}
