package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/cybersaksham/gogo/conf"
)

// RunserverConfig contains resolved settings for the development server.
type RunserverConfig struct {
	Addr         string
	SettingsPath string
	Reload       bool
	Settings     conf.Settings
}

// ServerStarter starts an HTTP server with the resolved runserver config.
type ServerStarter func(context.Context, RunserverConfig) error

// NewRunserverCommand creates the runserver command.
func NewRunserverCommand(starter ServerStarter) Command {
	if starter == nil {
		starter = unavailableServerStarter
	}

	return runserverCommand{starter: starter}
}

type runserverCommand struct {
	starter ServerStarter
}

func (c runserverCommand) Name() string {
	return "runserver"
}

func (c runserverCommand) Summary() string {
	return "Run the development server"
}

func (c runserverCommand) Run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("runserver", flag.ContinueOnError)
	addr := flags.String("addr", "", "address to bind")
	settingsPath := flags.String("settings", "", "settings file path")
	reload := flags.Bool("reload", false, "enable development reload")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidArguments, err)
	}

	settings, err := loadRunserverSettings(*settingsPath)
	if err != nil {
		return err
	}
	if err := settings.Validate(); err != nil {
		return err
	}

	resolvedAddr := settings.HTTPAddr
	if *addr != "" {
		resolvedAddr = *addr
	}

	return c.starter(ctx, RunserverConfig{
		Addr:         resolvedAddr,
		SettingsPath: *settingsPath,
		Reload:       *reload,
		Settings:     settings,
	})
}

func loadRunserverSettings(settingsPath string) (conf.Settings, error) {
	if settingsPath == "" {
		return conf.LoadFromEnv()
	}

	values, err := conf.LoadEnvFile(settingsPath)
	if err != nil {
		return conf.Settings{}, err
	}

	return conf.SettingsFromMap(values), nil
}

func unavailableServerStarter(context.Context, RunserverConfig) error {
	return fmt.Errorf("%w: runserver HTTP runtime is planned for 03-http-routing-middleware-views", ErrCommandUnavailable)
}
