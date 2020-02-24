//go:generate go run ./codegen/codegen.go
//go:generate go run ./vendor/github.com/go-bindata/go-bindata/go-bindata -o ./data/bindata.go -ignore bindata.go -pkg data -modtime 1557785965 -mode 0644 ./data/
package main

import (
	"io/ioutil"
	"os"
	"regexp"

	"github.com/mattn/go-colorable"
	"github.com/rancher/rke/cmd"
	"github.com/rancher/rke/metadata"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// VERSION gets overridden at build time using -X main.VERSION=$VERSION
var VERSION = "dev"
var released = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

func main() {
	logrus.SetOutput(colorable.NewColorableStdout())

	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {
	app := cli.NewApp()
	app.Name = "rke"
	app.Version = VERSION
	app.Usage = "Rancher Kubernetes Engine, an extremely simple, lightning fast Kubernetes installer that works everywhere"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if ctx.GlobalBool("quiet") {
			logrus.SetOutput(ioutil.Discard)
		}
		if released.MatchString(app.Version) {
			metadata.RKEVersion = app.Version
			return nil
		}
		logrus.Warnf("This is not an officially supported version (%s) of RKE. Please download the latest official release at https://github.com/rancher/rke/releases/latest", app.Version)
		return nil
	}
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Commands = []cli.Command{
		cmd.UpCommand(),
		cmd.RemoveCommand(),
		cmd.VersionCommand(),
		cmd.ConfigCommand(),
		cmd.EtcdCommand(),
		cmd.CertificateCommand(),
		cmd.EncryptionCommand(),
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug logging",
		},
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Quiet mode, disables logging and only critical output will be printed",
		},
	}
	return app.Run(os.Args)
}
