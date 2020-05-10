package command

import (
	"context"
	"os"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

func (m *Command) run() error {
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
