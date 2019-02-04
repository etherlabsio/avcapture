package main

import (
	"time"

	"github.com/etherlabsio/avcapture/internal/chrome"
	"github.com/etherlabsio/avcapture/internal/ffmpeg"

	execute "github.com/etherlabsio/avcapture/internal/exec"
	"github.com/etherlabsio/avcapture/pkg/recorder"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

const (
	waitTimeAfterChromeStartSec = 5
)

func handleRequest(req *recorder.StartRecordingRequest, appCtxt *appContext, status Status) error {
	var err error
	appCtxt.recordingLock.Lock()
	defer appCtxt.recordingLock.Unlock()
	switch status {
	case recordingStarted:
		err = handleStartedEvent(req, appCtxt)
	case recordingEnded:
		err = handleEndedEvent(appCtxt)
	default:
		err = errors.Errorf("recorder invalid event %s", status)
	}
	return err
}

func handleStartedEvent(req *recorder.StartRecordingRequest, appCtxt *appContext) (err error) {

	if appCtxt.isRunning {
		return errors.New("recoder: already running")
	}
	if 0 == len(req.FFmpeg.Params) {
		return errors.New("recorder: missing output params for encoding(ffmpeg)")
	}
	if "" == req.Chrome.URL {
		return errors.New("recorder: missing URL for capturing")
	}

	if 0 != len(req.FFmpeg.Options) {
		appCtxt.ff = appCtxt.ff.WithOptions(req.FFmpeg.Options...)
	}
	if 0 != len(req.Chrome.Options) {
		appCtxt.cr = appCtxt.cr.WithOptions(req.Chrome.Options...)
	}

	/* start chrome */
	{
		cmd, err := appCtxt.cr.WithURL(req.Chrome.URL).Build()
		if err != nil {
			return errors.Wrapf(err, "recorder: failed to build chrome cmd")
		}

		appCtxt.chromeCmd, err = chrome.Execute(appCtxt.logger, cmd...)
		if err != nil {
			return errors.Wrapf(err, "recorder: failed to start chrome cmd %s ", cmd)
		}
	}

	//TODO : Xvfb takes some time to allocate buffers in the beginning.
	// Adding extra delay, so that audio and video doesn't go out of sync.
	// Before meeting starts, if we try writing some dummy streams to Xvfb then it would have allocated buffer.
	// Then once meeting starts, this delay might be reduced.
	time.Sleep(waitTimeAfterChromeStartSec * time.Second)

	/* start ffmpeg */
	{
		cmd, err := appCtxt.ff.WithArguments(req.FFmpeg.Params...).Build()
		if err != nil {
			return errors.Wrapf(err, "recorder: failed to build ffmpeg cmd")
		}
		appCtxt.ffmpegCmd, err = ffmpeg.Execute(appCtxt.logger, cmd...)
		if err != nil {
			return errors.Wrapf(err, "recorder: failed to start ffmpeg cmd %s ", cmd)
		}
	}
	appCtxt.isRunning = true

	appCtxt.recSTime = time.Now().UTC()
	return err
}
func handleEndedEvent(appCtxt *appContext) (err error) {
	stopTime := time.Now().UTC()

	if !appCtxt.isRunning {
		return errors.New("recorder: is not running")
	}

	/* stop ffmpeg */
	{
		err = execute.StopCommand(appCtxt.logger, appCtxt.ffmpegCmd, false)
		if nil != err {
			level.Error(appCtxt.logger).Log("error", "stop-ffmpeg", "error-stack", err)
		}
	}

	/* stop chrome */
	{
		err = execute.StopCommand(appCtxt.logger, appCtxt.chromeCmd, false)
		if nil != err {
			level.Error(appCtxt.logger).Log("error", "stop-chrome", "error-stack", err)
		}
	}

	appCtxt.recETime = stopTime
	appCtxt.isRunning = false
	err = nil
	return err
}
