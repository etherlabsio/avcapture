package commander

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/etherlabsio/errors"
)

// Runnable provides a command than can be run and stopped
type Runnable interface {
	Start() error
	Stop() error
}

// Exec executes a shell command
func Exec(arg string) error {
	str := buildExecutableStr(arg)
	cmd := exec.Command(str)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WithMessagef(err, "exec: failed to pipe RunnableCmd %s", arg)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	fmt.Println("executing")
	err = errors.
		Do(cmd.Start).
		Do(cmd.Wait).
		Err()
	return errors.WithMessagef(err, "exec: failed to execute RunnableCmd %s", arg)
}

// RunnableCmd returns a runnable Shell command
type RunnableCmd struct {
	args string
	cmd  *exec.Cmd
}

func New(args ...string) *RunnableCmd {
	execStr := buildExecutableStr(args...)
	cmd := exec.Command(execStr)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return &RunnableCmd{execStr, cmd}
}

func (s *RunnableCmd) Start() error {
	return errors.WithMessage(s.cmd.Start(), "failed to start RunnableCmd: "+s.args)
}

func (s *RunnableCmd) Stop() error {
	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err != nil {
		return errors.Wrapf(err, "execute-command: failed to get process group id")
	}

	stop := func(pgid int, signal syscall.Signal) error {
		if err := syscall.Kill(-pgid, signal); err != nil {
			return errors.WithMessage(err, "execute-RunnableCmd: failed to terminate process group id")
		}
		if err := s.cmd.Wait(); err != nil {
			return errors.WithMessage(err, "execute-RunnableCmd: failed while waiting")
		}
		return nil
	}
	return stop(pgid, syscall.SIGTERM)
}

func buildExecutableStr(args ...string) string {
	args = append([]string{"sh", "-c"}, args...)
	fmt.Println("exec: ", args)
	return strings.Join(args, " ")
}
