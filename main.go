package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/haleyrc/vault/vault"
)

func main() {
	log.SetHandler(cli.New(os.Stderr))
	// log.SetLevel(log.DebugLevel)

	app, err := vault.NewApp()
	if err != nil {
		panic(err)
	}

	log.WithField("config_log", app.ConfigLoc).Debug("loaded configuration")

	if len(os.Args) < 2 {
		log.Fatal("add the service runner here")

		// Early return prevents the interactive stuff from running
		return
	}

	command := os.Args[1]
	args := os.Args[2:]
	code := app.Run(command, args...)
	os.Exit(code)
}
