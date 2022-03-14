package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rancher/rke/metadata"
)

const (
	defaultURL = "https://raw.githubusercontent.com/rancher/kontainer-driver-metadata/42cc26fd30453d6e2b9880d6508ea6004a4dd14e/data/data.json"
	dataFile   = "data/data.json"
)

// Codegen fetch data.json from https://releases.rancher.com/kontainer-driver-metadata/dev-v2.6/data.json and generates bindata
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

	b, err := ioutil.ReadAll(data.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Writing data")
	if err := ioutil.WriteFile(dataFile, b, 0755); err != nil {
		return
	}
}
