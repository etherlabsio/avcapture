package main

import (
	"os/exec"
	"sync"
	"time"

	"github.com/etherlabsio/avcapture/internal/chrome"
	"github.com/etherlabsio/avcapture/internal/ffmpeg"
	"github.com/go-kit/kit/log"
)

type ID string

type Status string

const (
	recordingStarted Status = "started"
	recordingEnded   Status = "ended"
)

type appContext struct {
	logger        log.Logger
	ff            ffmpeg.Builder
	cr            chrome.Builder
	recordingLock sync.Mutex
	recSTime      time.Time
	recETime      time.Time
	ffmpegCmd     *exec.Cmd
	chromeCmd     *exec.Cmd
	isRunning     bool
}
