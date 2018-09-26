package main

import (
	"os"
	"regexp"

	"github.com/mattn/go-colorable"
	"github.com/rancher/rke/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var VERSION = "v0.0.12-dev"
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
		logrus.Debugf("RKE version %s", app.Version)
		if released.MatchString(app.Version) {
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
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug logging",
		},
	}
	return app.Run(os.Args)
}
