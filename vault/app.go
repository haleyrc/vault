package vault

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/apex/log"
)

const (
	OK    = 0
	Error = 1
	Warn  = 2
)

func NewApp() (*App, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("new app: %w", err)
	}
	app := &App{
		ConfigLoc: filepath.Join(configDir, "vault", "config.json"),
		Config:    &Config{},
	}
	if err := app.loadConfig(); err != nil {
		return nil, fmt.Errorf("new app: %w", err)
	}
	return app, nil
}

type App struct {
	ConfigLoc string
	Config    *Config
}

func (a *App) Run(command string, args ...string) int {
	switch command {
	case "config":
		return a.handleConfig(args...)
	case "help":
		a.printUsage()
		return OK
	default:
		a.printUsage()
		return Error
	}
}

func (a *App) handleConfig(args ...string) int {
	if len(args) < 1 {
		a.printConfigUsage()
		return Error
	}
	command := args[0]
	switch command {
	case "adddir":
		return a.handleAddDir(args[1:]...)
	case "list":
		return a.handleListConfig(args[1:]...)
	case "help":
		a.printConfigUsage()
		return OK
	default:
		a.printConfigUsage()
		return Error
	}
}

func (a *App) handleListConfig(args ...string) int {
	fmt.Println("=== Shares")
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	for _, share := range a.Config.Shares {
		line := fmt.Sprintf("%s:\t%s\n", share.Name, share.Dir)
		tw.Write([]byte(line))
	}
	tw.Flush()
	return OK
}

func (a *App) handleAddDir(args ...string) int {
	var name string
	var dir string

	fs := flag.NewFlagSet("adddir", flag.ContinueOnError)
	fs.StringVar(&name, "name", "", "The name of the share to create")
	fs.StringVar(&dir, "dir", "", "The directory to backup to the share")
	fs.Parse(args)

	if name == "" || dir == "" {
		a.printAddDirUsage()
		return Error
	}

	a.Config.AddShare(name, dir)

	if err := a.saveConfig(); err != nil {
		log.WithError(err).Error("failed to save config")
		return Error
	}

	return OK
}

func (a *App) loadConfig() error {
	log.WithField("loc", a.ConfigLoc).Debug("loading config")

	f, err := os.Open(a.ConfigLoc)

	// If there's no config file, we just use the empty config
	if os.IsNotExist(err) {
		log.Debug("no config file found")
		return nil
	}

	// But if we just couldn't read it, we quit to prevent any issues
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	defer f.Close()

	log.Debug("parsing config json")
	if err := json.NewDecoder(f).Decode(a.Config); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return nil
}

func (a *App) saveConfig() error {
	dir := filepath.Dir(a.ConfigLoc)
	if err := mkdir(dir); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	f, err := os.Create(a.ConfigLoc)
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	if err := enc.Encode(a.Config); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

func (a *App) printAddDirUsage() {
	usage := `
vault config adddir --name VALUE --dir VALUE

Calling adddir adds a new directory to backup. The provided name will be used as
the share name, which corresponds to a top-level "folder" on S3. The directory
provided will be backed up recursively, preserving the file names and orders
under the top-level share.
`
	fmt.Println(strings.TrimSpace(usage))
}

func (a *App) printConfigUsage() {
	usage := `
vault config COMMAND [OPTION...]

Running vault with the config command allows you to view and/or modify the app
configuration using one of the subcommands below.

COMMAND:
    set     Set a new configuration value
    help    Print this help
`
	fmt.Println(strings.TrimSpace(usage))
}

func (a *App) printUsage() {
	usage := `
vault [COMMAND] [OPTION...]

Running vault with no command starts the app in service mode. In this mode, the
backup process will be performed periodically in a loop. This is how the Windows
Service Controller starts up the app. To run in interactive mode, one of the
commands below must be present.

COMMAND:
    config    View or modify the app configuration
    help      Print this help
`
	fmt.Println(strings.TrimSpace(usage))
}

func mkdir(dir string) error {
	fi, err := os.Stat(dir)
	if err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("mkdir: not a directory")
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("mkdir: %w", err)
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return nil
}
