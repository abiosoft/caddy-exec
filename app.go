package command

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// Interface guards
var (
	_ caddy.App             = (*App)(nil)
	_ caddy.Module          = (*App)(nil)
	_ caddy.Provisioner     = (*App)(nil)
	_ caddy.Validator       = (*App)(nil)
	_ caddyfile.Unmarshaler = (*App)(nil)
)

func init() {
	caddy.RegisterModule(App{})
}

// App is top level module that runs shell commands.
type App struct {
	moduleConfig
}

// Start starts the app.
func (a *App) Start() error {
	if !a.Startup {
		return nil
	}

	return a.run()
}

// Stop stops the app.
func (a *App) Stop() error {
	if !a.Shutdown {
		return nil
	}

	return a.run()
}

// Provision implements caddy.Provisioner.
func (a *App) Provision(ctx caddy.Context) error {
	if a.empty() {
		// special case
		// provision will be manually called
		// workaround to take advantage of caddyfile
		return nil
	}

	err := a.moduleConfig.provision(ctx, a)
	if err != nil {
		return err
	}

	if !a.Startup && !a.Shutdown {
		return errors.New("one of startup|shutdown is required")
	}

	if a.Startup && a.Shutdown {
		return errors.New("only one of startup|shutdown is required")
	}

	return nil
}

// Validate implements caddy.Validator.
func (a App) Validate() error {
	if a.moduleConfig.empty() {
		// special case
		return nil
	}

	return a.moduleConfig.Validate()
}

// CaddyModule implements caddy.ModuleInfo
func (a App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "exec",
		New: func() caddy.Module { return new(App) },
	}
}
