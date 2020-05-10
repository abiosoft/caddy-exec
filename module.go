package command

import (
	"fmt"
	"os"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

// Interface guards
var (
	_ caddy.Provisioner           = (*Exec)(nil)
	_ caddy.Validator             = (*Exec)(nil)
	_ caddyhttp.MiddlewareHandler = (*Handler)(nil)
	_ caddyfile.Unmarshaler       = (*Exec)(nil)
)

func init() {
	caddy.RegisterModule(Handler{})
	httpcaddyfile.RegisterHandlerDirective("exec", parseHandlerCaddyfile)
}

// Exec module implements an HTTP handler that runs a shell command.
type Exec struct {
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
func (Exec) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "exec",
		New: func() caddy.Module { return new(Exec) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Exec) Provision(ctx caddy.Context) error {
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
func (m Exec) Validate() error {
	if m.Command == "" {
		return fmt.Errorf("command is required")
	}

	if err := isValidDir(m.Directory); err != nil {
		return err
	}
	return nil
}

func isValidDir(dir string) error {
	// current directory is valid
	if dir == "" {
		return nil
	}

	s, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !s.IsDir() {
		return fmt.Errorf("not a directory '%s'", dir)
	}
	return nil
}
