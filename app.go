package command

import (
	"io"
	"sync/atomic"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

// Interface guards
var (
	_ caddy.App         = (*App)(nil)
	_ caddy.Module      = (*App)(nil)
	_ caddy.Provisioner = (*App)(nil)
	_ caddy.Validator   = (*App)(nil)
)

// lifeCycle is used to keep track of startup/shutdown
var lifeCycle int32

// loggers keeps track of loggers to prevent recreation.
var loggers = map[string]io.WriteCloser{}

func init() {
	caddy.RegisterModule(App{})
}

// App is top level module that runs shell commands.
type App struct {
	Commands []Cmd `json:"commands,omitempty"`

	commands map[string][]Runner
	log      *zap.Logger
}

// Provision implements caddy.Provisioner
func (a *App) Provision(ctx caddy.Context) error {
	if a.commands == nil {
		a.commands = map[string][]Runner{}
	}

	a.log = ctx.Logger(a)
	for _, cmd := range a.Commands {
		if err := cmd.provision(ctx, a); err != nil {
			return err
		}
		a.addCmd(cmd)
	}
	return nil
}

func (a *App) addCmd(c Cmd) {
	runner := runnerFunc(c.run)
	for at := range c.at {
		a.commands[at] = append(a.commands[at], runner)
	}
}

// Validate implements caddy.Validator
func (a App) Validate() error {
	for _, cmd := range a.Commands {
		if err := cmd.validate(); err != nil {
			return err
		}
	}

	return nil
}

// Start starts the app.
func (a App) Start() error {
	count := atomic.AddInt32(&lifeCycle, 1)
	if count > 1 {
		// not the first startup, maybe a reload
		return nil
	}

	for _, runner := range a.commands["startup"] {
		if err := runner.Run(); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the app.
func (a *App) Stop() error {
	count := atomic.AddInt32(&lifeCycle, -1)
	if count > 0 {
		// not shutdown, maybe a prior config reload.
		return nil
	}

	for _, runner := range a.commands["shutdown"] {
		if err := runner.Run(); err != nil {
			return err
		}
	}
	return nil
}

// CaddyModule implements caddy.ModuleInfo
func (a App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "exec",
		New: func() caddy.Module { return new(App) },
	}
}
