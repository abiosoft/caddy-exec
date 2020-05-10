package command

import (
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

// Interface guards
var (
	_ caddy.Provisioner           = (*Command)(nil)
	_ caddy.Validator             = (*Command)(nil)
	_ caddyhttp.MiddlewareHandler = (*Command)(nil)
	_ caddyfile.Unmarshaler       = (*Command)(nil)
)

func init() {
	caddy.RegisterModule(Command{})
	httpcaddyfile.RegisterHandlerDirective("command", parseCaddyfile)
}

// Command module implements an HTTP handler that runs a shell command.
type Command struct {
	// The command to run.
	Command string `json:"command,omitempty"`
	// The command args.
	Args []string `json:"args,omitempty"`
	// The directory to run the command from.
	// Defaults to current directory.
	Directory string `json:"directory,omitempty"`
	// If the command should run in the foreground.
	// Setting it makes the command run in the foreground.
	Foreground bool `json:"foreground,omitempty"`

	// Timeout for the command. The command will be killed
	// after timeout has elapsed if it is still running.
	// Defaults to 10s.
	Timeout string `json:"timeout,omitempty"`

	timeout time.Duration // for ease of use after parsing
	log     *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (Command) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.command",
		New: func() caddy.Module { return new(Command) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Command) Provision(ctx caddy.Context) error {
	m.log = ctx.Logger(m)

	if m.Timeout == "" {
		m.Timeout = "10s"
	}

	if m.Timeout != "" {
		dur, err := time.ParseDuration(m.Timeout)
		if err != nil {
			return err
		}
		m.timeout = dur
	}

	return nil
}

// Validate implements caddy.Validator.
func (m Command) Validate() error {
	if m.Command == "" {
		return fmt.Errorf("command is required")
	}

	if err := isValidDir(m.Directory); err != nil {
		return err
	}
	return nil
}
