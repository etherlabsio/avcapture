package chrome

import (
	"os/exec"
	"strings"

	"github.com/go-kit/kit/log"

	execute "github.com/etherlabsio/avcapture/internal/exec"
)

type Builder struct {
	options []string
	URL     string
}

func NewBuilder() Builder {
	return Builder{
		options: []string{"--enable-logging=stderr", "--autoplay-policy=no-user-gesture-required", "--no-sandbox", "--disable-infobars", "--kiosk", "--start-maximized --window-position=0,0", "--window-size=1280,720"},
	}
}

func (b Builder) WithOptions(options ...string) Builder {
	b.options = options
	return b
}

func (b Builder) WithURL(URL string) Builder {
	b.URL = URL
	return b
}

func (b Builder) Build() ([]string, error) {
	var cmdArgs []string

	cmdArgs = append(cmdArgs, b.options...)
	if b.URL != "" {
		cmdArgs = append(cmdArgs, "--app=\""+b.URL+"\"")

	}
	return cmdArgs, nil

}

func Execute(logger log.Logger, cmdArgs ...string) (*exec.Cmd, error) {
	var cmd []string
	var envs []string
	cmd = append(cmd, "/usr/bin/google-chrome-stable")
	cmd = append(cmd, cmdArgs...)
	envs = append(envs, "DISPLAY=:99")

	return execute.StartCommand(logger, strings.Join(cmd, " "), envs, true)
}
