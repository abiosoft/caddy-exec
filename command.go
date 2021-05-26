package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

// Cmd is the module configuration
type Cmd struct {
	// The command to run.
	Command string `json:"command,omitempty"`

	// The command args.
	Args []string `json:"args,omitempty"`

	// The directory to run the command from.
	// Defaults to current directory.
	Directory string `json:"directory,omitempty"`

	// If the command should run in the foreground.
	// By default, commands run in the background and doesn't
	// affect Caddy.
	// Setting this makes the command run in the foreground.
	// Note that failure of a startup command running in the
	// foreground may prevent Caddy from starting.
	Foreground bool `json:"foreground,omitempty"`

	// Timeout for the command. The command will be killed
	// after timeout has elapsed if it is still running.
	// Defaults to 10s.
	Timeout string `json:"timeout,omitempty"`

	// When the command should run. This can contain either of
	// "startup" or "shutdown".
	At []string `json:"at,omitempty"`

	// Standard output log.
	StdWriterRaw json.RawMessage `json:"log,omitempty" caddy:"namespace=caddy.logging.writers inline_key=output"`

	// Standard error log.
	ErrWriterRaw json.RawMessage `json:"err_log,omitempty" caddy:"namespace=caddy.logging.writers inline_key=output"`

	timeout time.Duration       // ease of use after parsing timeout string
	at      map[string]struct{} // for quicker access and uniqueness.
	log     *zap.Logger

	// logging
	stdWriter io.WriteCloser
	errWriter io.WriteCloser
}

// Provision implements caddy.Provisioner.
func (c *Cmd) provision(ctx caddy.Context, cm caddy.Module) error {
	c.log = ctx.Logger(cm)

	// timeout
	if c.Timeout == "" {
		c.Timeout = "10s"
	}
	dur, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return err
	}
	c.timeout = dur

	// at
	if c.at == nil {
		c.at = map[string]struct{}{}
	}
	for _, at := range c.At {
		c.at[at] = struct{}{}
	}

	// std logger
	c.stdWriter, err = writerFromRaw(ctx, c, "StdWriterRaw", c.StdWriterRaw)
	if err != nil {
		return err
	}
	// err logger is optional
	if c.ErrWriterRaw != nil {
		c.errWriter, err = writerFromRaw(ctx, c, "ErrWriterRaw", c.ErrWriterRaw)
		if err != nil {
			return err
		}
	}

	return nil
}

func writerFromRaw(ctx caddy.Context, c *Cmd, field string, w json.RawMessage) (io.WriteCloser, error) {
	var err error
	var writerOpener caddy.WriterOpener
	var writer io.WriteCloser
	if w != nil {
		mod, err := ctx.LoadModule(c, field) //"WriterRaw"
		if err != nil {
			return nil, fmt.Errorf("loading log writer module: %v", err)
		}
		writerOpener = mod.(caddy.WriterOpener)
	}
	if writerOpener == nil {
		writerOpener = caddy.StderrWriter{}
	}

	if w, ok := loggers[writerOpener.WriterKey()]; ok {
		writer = w
	} else {
		writer, err = writerOpener.OpenWriter()
		if err != nil {
			return nil, fmt.Errorf("opening log writer using %#v: %v", writerOpener, err)
		}
	}

	return writer, err
}

// Validate implements caddy.Validator.
func (c Cmd) validate() error {
	if c.Command == "" {
		return fmt.Errorf("command is required")
	}

	if err := isValidDir(c.Directory); err != nil {
		return err
	}

	for _, at := range c.At {
		switch at {
		case "startup":
		case "shutdown":
		default:
			return fmt.Errorf("'at' can only contain one of 'startup' or 'shutdown'")
		}
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
