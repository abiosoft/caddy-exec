package command

import (
	"context"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

// Runner runs a command.
type Runner interface {
	Run() error
}

type runnerFunc func() error

func (r runnerFunc) Run() error { return r() }

func (c *Cmd) run(args []string) error {
	cmdInfo := zap.Any("command", append([]string{c.Command}, args...))
	log := c.log.With(cmdInfo)
	startTime := time.Now()

	cmd := exec.Command(c.Command, args...)

	done := make(chan struct{}, 1)

	// timeout
	if c.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)

		// the context must not be cancelled before the command is done
		go func() {
			<-done
			cancel()
		}()

		cmd = exec.CommandContext(ctx, c.Command, args...)
	}

	// configure command
	{
		cmd.Stdout = c.stdWriter
		cmd.Stderr = c.stdWriter
		if c.errWriter != nil {
			cmd.Stderr = c.errWriter
		}
		cmd.Dir = c.Directory
	}

	wait := func(err error) error {
		// only wait if start was successful
		if cmd.Process != nil {
			// err is empty, we can reuse it without losing any info
			err = cmd.Wait()
		}
		done <- struct{}{}

		log = log.With(zap.Duration("duration", time.Since(startTime))).Named("exit")

		if err != nil {
			log.Error("", zap.Error(err))
			return err
		}

		log.Info("")
		return nil
	}

	// start command
	err := cmd.Start()

	if c.Foreground {
		return wait(err)
	}

	go wait(err)
	return err
}
