package exec

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

func ExecuteCommand(logger log.Logger, cmd string) (string, error) {
	level.Info(logger).Log("info", "executing-command", "command", cmd)
	cmdBuilder := exec.Command("sh", "-c", cmd)
	stdout, err := cmdBuilder.StdoutPipe()
	if err != nil {
		return "", errors.Wrapf(err, "execute-command: failed to pipe command %s", cmd)
	}
	if err := cmdBuilder.Start(); err != nil {
		return "", errors.Wrapf(err, "execute-command: failed to start command %s", cmd)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	if err := cmdBuilder.Wait(); err != nil {
		return "", errors.Wrapf(err, "execute-command: failed to execute command %s", cmd)
	}

	return buf.String(), nil
}

func StartCommand(logger log.Logger, cmd string, envs []string, logToStd bool) (*exec.Cmd, error) {
	level.Info(logger).Log("info", "starting-command", "command", cmd)
	cmdBuilder := exec.Command("sh", "-c", cmd)
	cmdBuilder.Env = append(os.Environ(), envs...)
	cmdBuilder.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if logToStd {
		cmdBuilder.Stdout = os.Stdout
		cmdBuilder.Stderr = os.Stderr
	}

	if err := cmdBuilder.Start(); err != nil {
		return cmdBuilder, errors.Wrapf(err, "execute-command: failed to start command %s", cmd)
	}

	return cmdBuilder, nil
}

func StopCommand(logger log.Logger, cmdBuilder *exec.Cmd, kill bool) error {
	level.Info(logger).Log("info", "stopping-command", "command", cmdBuilder.Path)
	pgid, err := syscall.Getpgid(cmdBuilder.Process.Pid)
	if err != nil {
		return errors.Wrapf(err, "execute-command: failed to get process group id")
	}

	signal := syscall.SIGTERM
	if kill {
		signal = syscall.SIGKILL
	}

	if err := syscall.Kill(-pgid, signal); err != nil {
		return errors.Wrapf(err, "execute-command: failed to terminate process group id")
	}

	if err := cmdBuilder.Wait(); err != nil {
		return errors.Wrapf(err, "execute-command: failed while waiting")
	}

	level.Info(logger).Log("info", "stopping-done", "command-state", cmdBuilder.Process.Pid)

	return nil
}
