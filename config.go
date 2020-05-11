package command

import (
	"fmt"
	"os"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

// moduleConfig is the module configuration
type moduleConfig struct {
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
	Timeout string        `json:"timeout,omitempty"`
	timeout time.Duration // for ease of use after parsing

	// If the command should run at startup.
	Startup bool `json:"startup,omitempty"`
	// If the command should run at shutdown.
	Shutdown bool `json:"shutdown,omitempty"`

	// run at startup/shutdown
	handler bool
	log     *zap.Logger
}

// empty returns if the config is empty.
func (m moduleConfig) empty() bool {
	switch {
	case m.Command != "":
	case m.Args != nil:
	case m.Directory != "":
	case m.Timeout != "":
	case m.Foreground:
	case m.Startup:
	case m.Shutdown:
	default:
		return true
	}
	return false
}

// Provision implements caddy.Provisioner.
func (m *moduleConfig) provision(ctx caddy.Context, cm caddy.Module) error {
	m.log = ctx.Logger(cm)

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
func (m moduleConfig) Validate() error {
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
