package ffmpeg

import (
	"os/exec"
	"strings"

	"github.com/go-kit/kit/log"

	execute "github.com/etherlabsio/avcapture/internal/exec"
)

type Builder struct {
	options [][]string
	args    [][]string
}

func NewBuilder() Builder {
	return Builder{
		options: [][]string{{"-y", ""}, {"-v", "info"}, {"-f", "x11grab"}, {"-draw_mouse", "0"}, {"-r", "24"}, {"-s", "1280x720"}, {"-thread_queue_size", "4096"}, {"-i", ":99.0+0,0"},
			{"-f", "pulse"}, {"-thread_queue_size", "4096"}, {"-i", "default"},
			{"-acodec", "aac"}, {"-strict", "-2"}, {"-ar", "44100"},
			{"-c:v", "libx264"}, {"-x264opts", "no-scenecut"}, {"-preset", "veryfast"}, {"-profile:v", "main"}, {"-level", "3.1"}, {"-pix_fmt", "yuv420p"},
			{"-r", "24"}, {"-crf", "25"}, {"-g", "48"}, {"-keyint_min", "48"},
			{"-force_key_frames", "\"expr:gte(t,n_forced*2)\""}, {"-tune", "zerolatency"}, {"-b:v", "2800k"}, {"-maxrate", "2996k"}, {"-bufsize", "4200k"}},
	}
}

func (b Builder) WithOptions(options ...[]string) Builder {
	b.options = options
	return b
}

func (b Builder) WithArguments(args ...[]string) Builder {
	b.args = args
	return b
}

func (b Builder) Build() ([]string, error) {
	var cmdArgs []string

	for _, p := range b.options {
		cmdArgs = append(cmdArgs, p...)
	}
	for _, p := range b.args {
		cmdArgs = append(cmdArgs, p...)
	}
	return cmdArgs, nil

}

func Execute(logger log.Logger, cmdArgs ...string) (*exec.Cmd, error) {
	var cmd []string
	var envs []string
	cmd = append(cmd, "/usr/bin/ffmpeg")
	cmd = append(cmd, cmdArgs...)
	return execute.StartCommand(logger, strings.Join(cmd, " "), envs, true)
}
