package command

import (
	"context"
	"os"
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

func (m *Cmd) run() error {
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

	if m.Foreground {
		return wait(err)
	}

	go wait(err)
	return err
}
