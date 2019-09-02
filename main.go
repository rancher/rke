package main

import (
	"github.com/rancher/rke/metadata"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/rancher/rke/cmd"
	"github.com/rancher/rke/util"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// VERSION gets overridden at build time using -X main.VERSION=$VERSION
var VERSION = "dev"
var released = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
var proxyEnvVars = [3]string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"}

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
		// Print proxy related environment variables
		for _, proxyEnvVar := range proxyEnvVars {
			var err error
			// Lookup environment variable
			if key, value, ok := util.GetEnvVar(proxyEnvVar); ok {
				// If it can contain a password, strip it (HTTP_PROXY or HTTPS_PROXY)
				if strings.HasPrefix(strings.ToUpper(proxyEnvVar), "HTTP") {
					value, err = util.StripPasswordFromURL(value)
					if err != nil {
						// Don't error out of provisioning when parsing of environment variable fails
						logrus.Warnf("Error parsing proxy environment variable %s", key)
						continue
					}
				}
				logrus.Infof("Using proxy environment variable %s with value [%s]", key, value)
			}
		}
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
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Quiet mode, disables logging and only critical output will be printed",
		},
	}
	return app.Run(os.Args)
}
