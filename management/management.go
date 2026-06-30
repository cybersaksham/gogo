package management

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cybersaksham/gogo/app"
	"github.com/cybersaksham/gogo/conf"
	gogohttp "github.com/cybersaksham/gogo/http"
	"github.com/cybersaksham/gogo/internal/cli"
	"github.com/cybersaksham/gogo/models"
	"github.com/cybersaksham/gogo/orm"
	"github.com/cybersaksham/gogo/queue"
)

// Project connects generated client project wiring to management commands.
type Project struct {
	Settings      func() conf.Settings
	AppConfigs    func() []app.Config
	ModelMetadata func() []models.Metadata
	Router        func() (*gogohttp.Router, error)
	QueueApp      func() *queue.App
}

// Execute runs the Gogo management command registry.
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	return cli.NewRoot().Execute(ctx, args, stdout, stderr)
}

// ExecuteProject runs management commands with generated project wiring.
func ExecuteProject(ctx context.Context, args []string, stdout, stderr io.Writer, project Project) error {
	return cli.NewRootWithOptions(cli.RootOptions{
		RunserverStarter: project.serverStarter(stdout),
		QueueRuntime:     project.queueRuntime(),
		FixtureStore:     project.fixtureStore(context.Background()),
	}).Execute(ctx, args, stdout, stderr)
}

// Main runs management commands using os.Args and exits with a process status.
func Main() {
	if err := Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// MainProject runs project-aware management commands using os.Args.
func MainProject(project Project) {
	if err := ExecuteProject(context.Background(), os.Args[1:], os.Stdout, os.Stderr, project); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (p Project) queueRuntime() *cli.QueueRuntime {
	if p.QueueApp == nil {
		return nil
	}
	runtime := cli.NewQueueRuntime()
	if app := p.QueueApp(); app != nil {
		runtime.App = app
	}
	return runtime
}

func (p Project) fixtureStore(ctx context.Context) cli.FixtureStore {
	if p.ModelMetadata == nil {
		return nil
	}
	settings, err := conf.LoadFromEnv()
	if err != nil {
		return cli.NewErrorFixtureStore(err)
	}
	if p.Settings != nil {
		settings = mergeSettings(p.Settings(), settings)
	}
	database, err := orm.OpenDatabaseURL(ctx, orm.DefaultDatabase, settings.DatabaseURL)
	if err != nil {
		return cli.NewErrorFixtureStore(err)
	}
	return cli.NewMetadataFixtureStore(orm.NewMetadataStore(database, p.ModelMetadata()...))
}

func (p Project) serverStarter(accessLog io.Writer) cli.ServerStarter {
	if p.Router == nil && p.AppConfigs == nil && p.Settings == nil {
		return nil
	}
	return func(ctx context.Context, config cli.RunserverConfig) error {
		settings := config.Settings
		if p.Settings != nil {
			settings = mergeSettings(p.Settings(), config.Settings)
		}
		settings.HTTPAddr = config.Addr

		registry := app.NewRegistry()
		if p.AppConfigs != nil {
			for _, config := range p.AppConfigs() {
				if err := registry.Register(config); err != nil {
					return err
				}
			}
		}

		router := gogohttp.NewRouter()
		if p.Router != nil {
			resolved, err := p.Router()
			if err != nil {
				return err
			}
			if resolved != nil {
				router = resolved
			}
		}

		middleware, err := gogohttp.BuildMiddleware(settings, gogohttp.BuiltInMiddlewareRegistry(accessLog))
		if err != nil {
			return err
		}

		server, err := gogohttp.NewServer(gogohttp.ServerConfig{
			Settings:   settings,
			Registry:   registry,
			Router:     router,
			Middleware: middleware,
		})
		if err != nil {
			return err
		}
		return server.ListenAndServe(ctx)
	}
}

func mergeSettings(projectSettings conf.Settings, loaded conf.Settings) conf.Settings {
	merged := projectSettings
	if loaded.Env != "" {
		merged.Env = loaded.Env
	}
	if loaded.SecretKey != "" {
		merged.SecretKey = loaded.SecretKey
	}
	merged.Debug = loaded.Debug
	if len(loaded.AllowedHosts) > 0 {
		merged.AllowedHosts = loaded.AllowedHosts
	}
	if loaded.HTTPAddr != "" {
		merged.HTTPAddr = loaded.HTTPAddr
	}
	if loaded.DatabaseURL != "" {
		merged.DatabaseURL = loaded.DatabaseURL
	}
	if len(loaded.InstalledApps) > 0 {
		merged.InstalledApps = loaded.InstalledApps
	}
	if len(loaded.Middleware) > 0 {
		merged.Middleware = loaded.Middleware
	}
	if loaded.RootURLConf != "" && loaded.RootURLConf != conf.DefaultSettings().RootURLConf {
		merged.RootURLConf = loaded.RootURLConf
	}
	if loaded.StaticURL != "" {
		merged.StaticURL = loaded.StaticURL
	}
	if loaded.StaticRoot != "" {
		merged.StaticRoot = loaded.StaticRoot
	}
	if loaded.MediaURL != "" {
		merged.MediaURL = loaded.MediaURL
	}
	if loaded.MediaRoot != "" {
		merged.MediaRoot = loaded.MediaRoot
	}
	if len(loaded.TemplateDirs) > 0 {
		merged.TemplateDirs = loaded.TemplateDirs
	}
	if loaded.DefaultAutoField != "" {
		merged.DefaultAutoField = loaded.DefaultAutoField
	}
	if loaded.TimeZone != "" {
		merged.TimeZone = loaded.TimeZone
	}
	if loaded.LanguageCode != "" {
		merged.LanguageCode = loaded.LanguageCode
	}
	if loaded.SessionCookieName != "" {
		merged.SessionCookieName = loaded.SessionCookieName
	}
	merged.SessionCookieSecure = loaded.SessionCookieSecure
	if loaded.CSRFCookieName != "" {
		merged.CSRFCookieName = loaded.CSRFCookieName
	}
	merged.CSRFCookieSecure = loaded.CSRFCookieSecure
	merged.HTTPSEnabled = loaded.HTTPSEnabled
	if len(loaded.CSRFTrustedOrigins) > 0 {
		merged.CSRFTrustedOrigins = loaded.CSRFTrustedOrigins
	}
	if loaded.AdminPath != "" {
		merged.AdminPath = loaded.AdminPath
	}
	merged.AdminPathReviewed = loaded.AdminPathReviewed
	merged.MigrationsApplied = loaded.MigrationsApplied
	merged.StaticFilesCollected = loaded.StaticFilesCollected
	merged.PasswordResetEnabled = loaded.PasswordResetEnabled
	if loaded.BrokerURL != "" {
		merged.BrokerURL = loaded.BrokerURL
	}
	if loaded.ResultBackend != "" {
		merged.ResultBackend = loaded.ResultBackend
	}
	if loaded.CacheURL != "" {
		merged.CacheURL = loaded.CacheURL
	}
	if loaded.EmailURL != "" {
		merged.EmailURL = loaded.EmailURL
	}
	return merged
}
