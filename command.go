package command

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("command", parseCaddyfile)
}

// Middleware implements an HTTP handler that runs a shell command.
type Middleware struct {
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
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.command",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	m.log = ctx.Logger(m)
	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	if m.Command == "" {
		return fmt.Errorf("command is required")
	}

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

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	var resp struct {
		Status string `json:"status,omitempty"`
		Error  string `json:"error,omitempty"`
	}
	err := m.run()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = err.Error()
	} else {
		resp.Status = "success"
	}

	return json.NewEncoder(w).Encode(resp)
}

// UnmarshalCaddyfile configures the plugin from Caddyfile.
// Syntax:
//
//   command [<matcher>] <command> [args...] {
//       args        <text>...
//       directory   <text>
//       timeout     <duration>
//       foreground
//   }
//
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Next() {
		return d.ArgErr()
	}

	if !d.Args(&m.Command) {
		return d.ArgErr()
	}
	m.Args = d.RemainingArgs()

	for d.NextBlock(0) {
		switch d.Val() {
		case "args":
			if len(m.Args) > 0 {
				return d.Err("args specified twice")
			}
			m.Args = d.RemainingArgs()
		case "directory":
			if !d.Args(&m.Directory) {
				return d.ArgErr()
			}
			m.Directory = d.Val()
			if err := isValidDir(m.Directory); err != nil {
				return err
			}
		case "foreground":
			m.Foreground = true
		case "timeout":
			if !d.Args(&m.Timeout) {
				return d.ArgErr()
			}
			dur, err := time.ParseDuration(m.Timeout)
			if err != nil {
				return err
			}
			m.timeout = dur
		}
	}

	return nil
}

func isValidDir(dir string) error {
	s, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !s.IsDir() {
		return fmt.Errorf("not a directory '%s'", dir)
	}
	return nil
}

func (m *Middleware) run() error {
	cmdInfo := zap.Any("command", append([]string{m.Command}, m.Args...))
	log := m.log.With(cmdInfo)
	startTime := time.Now()

	cmd := exec.Command(m.Command, m.Args...)

	done := make(chan struct{}, 1)

	// timeout
	if m.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), m.timeout)

		// the context must not be cancelled before the command is done
		go func() {
			<-done
			cancel()
		}()

		cmd = exec.CommandContext(ctx, m.Command, m.Args...)
	}

	// configure command
	{
		// TODO: improve logger
		writer := os.Stderr
		cmd.Stderr = writer
		cmd.Stdout = writer
		cmd.Dir = m.Directory
	}

	wait := func(err error) error {
		// only wait if start was successful
		if cmd.Process != nil {
			// err is empty, we can reuse it without losing any info
			err = cmd.Wait()
		}
		done <- struct{}{}

		duration := time.Since(startTime).String()
		log = log.With(zap.Any("duration", duration))

		if err != nil {
			log.Error("exit", zap.Any("error", err))
			return err
		}

		log.Info("exit", zap.Any("command", m.Command))
		return nil
	}

	// start command
	err := cmd.Start()

	if m.Foreground {
		return wait(err)
	}

	go wait(err)
	return err
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
