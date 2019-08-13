package recorder

import (
	"sync"
	"time"

	"github.com/etherlabsio/errors"
)

type Runnable interface {
	Start() error
	Stop() error
	Restart() error
}

type state int

const (
	idle state = 1 + iota
	running
	reloading
)

type Recorder struct {
	state               state
	lastHealthCheckedAt time.Time
	mtx                 sync.Mutex
	destination         string

	FFmpegCmd Runnable
	ChromeCmd Runnable
}

type FFmpeg struct {
	Params  [][]string `json:"params"`
	Options [][]string `json:"options"`
}

type Chrome struct {
	URL     string   `json:"url"`
	Options []string `json:"options"`
}

func cleanup(rec *Recorder) {
	rec.state = idle
	rec.FFmpegCmd = nil
	rec.ChromeCmd = nil
	rec.destination = ""
}

func setRunInfo(rec *Recorder, ffmpeg, chrome Runnable, destination string) {
	rec.ChromeCmd = chrome
	rec.FFmpegCmd = ffmpeg
	rec.state = running
	rec.destination = destination
}

func setIdle(rec *Recorder) {
	rec.state = idle
}

func setReloading(rec *Recorder) error {
	if rec.state != running {
		return errors.New("recorder not in running state")
	}
	rec.state = reloading
	return nil
}

func setRunning(rec *Recorder) error {
	if rec.state == running {
		return errors.New("recorder is already in running state")
	}
	rec.state = running
	return nil
}

func setRunningFromReloading(rec *Recorder) error {
	if rec.state != reloading {
		return errors.New("recorder is not in reloading state")
	}
	rec.state = running
	return nil
}
