package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rke/cmd"
	"github.com/urfave/cli"
)

var VERSION = "v0.1.0-dev"

func main() {
	app := cli.NewApp()
	app.Name = "rke"
	app.Version = VERSION
	app.Usage = "Rancher Kubernetes Engine, Running kubernetes cluster in the cloud"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Commands = []cli.Command{
		cmd.ClusterCommand(),
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug logging",
		},
	}
	app.Run(os.Args)
}
