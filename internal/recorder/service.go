package recorder

import (
	"context"
	"time"

	"github.com/etherlabsio/avcapture/pkg/chrome"
	"github.com/etherlabsio/avcapture/pkg/commander"
	"github.com/etherlabsio/avcapture/pkg/ffmpeg"
	"github.com/etherlabsio/errors"
	"github.com/go-kit/kit/log"
)

type errResponse struct {
	Err error `json:"error,omitempty"`
}

const (
	healthCheckInterval          = 2 * time.Second
	maxUnhealthyRecorderInterval = 5 * time.Second
	initHealthCheckWait          = 30 * time.Second
	reloadWaitInterval           = 2 * time.Second
)

func (e errResponse) Failed() error {
	if e.Err == nil {
		return nil
	}
	return e.Err
}

// StartRecordingRequest is the payload being received by recorder as part of start_recording
type StartRecordingRequest struct {
	FFmpeg `json:"ffmpeg"`
	Chrome `json:"chrome"`
}

// StartRecordingResponse defines response structure for the stop recording request
type StartRecordingResponse struct {
	StartTime time.Time `json:"start_time"`
	errResponse
}

type StopRecordingRequest struct{}

// StopRecordingResponse is the response for the stop recording recording request
type StopRecordingResponse struct {
	StopTime time.Time `json:"stop_time"`
	errResponse
}

type Service interface {
	Start(context.Context, StartRecordingRequest) StartRecordingResponse
	Stop(context.Context, StopRecordingRequest) StopRecordingResponse
	Check(context.Context) error
}

type service struct {
	recorder *Recorder
	logger   log.Logger
}

func NewService(l log.Logger) Service {
	return &service{
		recorder: &Recorder{state: idle},
		logger:   l,
	}
}

const (
	AlreadyRunning errors.Kind = iota + 5100
	AlreadyEnded
)

func (svc *service) Start(ctx context.Context, req StartRecordingRequest) (resp StartRecordingResponse) {
	const chromeLaunchWaitTime = 5 * time.Second
	svc.recorder.mtx.Lock()
	defer svc.recorder.mtx.Unlock()

	if svc.recorder.state == running {
		resp.Err = errors.New("recorder already running", AlreadyRunning)
		return resp
	}

	var ffmpegCmd Runnable
	{
		ffmpegBuilder := ffmpeg.NewBuilder()
		ffmpegBuilder = ffmpegBuilder.WithOptions(req.FFmpeg.Options...)
		ffmpegBuilder = ffmpegBuilder.WithArguments(req.FFmpeg.Params...)
		args, err := ffmpegBuilder.Build()
		if err != nil {
			resp.Err = errors.New("ffmpeg input invalid", errors.Invalid, err)
			return resp
		}
		ffmpegCmd = commander.New(args...)
	}

	var chromeCmd Runnable
	{
		chromeBuilder := chrome.NewBuilder()
		chromeBuilder = chromeBuilder.WithOptions(req.Chrome.Options...)
		chromeBuilder = chromeBuilder.WithURL(req.Chrome.URL)
		args, err := chromeBuilder.Build()
		if err != nil {
			resp.Err = errors.New("chrome input invalid", errors.Invalid, err)
			return resp
		}
		chromeCmd = commander.New(args...)
	}

	err := errors.
		Do(chromeCmd.Start).
		Do(func() error {
			//TODO : Xvfb takes some time to allocate buffers in the beginning.
			// Adding extra delay, so that audio and video doesn't go out of sync.
			// Before meeting starts, if we try writing some dummy streams to Xvfb then it would have allocated buffer.
			// Then once meeting starts, this delay might be reduced.
			time.Sleep(chromeLaunchWaitTime)
			return nil
		}).Do(ffmpegCmd.Start).
		Err()

	if err != nil {
		resp.Err = errors.New("failed to run the avcapture pipeline", err)
		return resp
	}

	setRunInfo(svc.recorder, ffmpegCmd, chromeCmd)
	resp.StartTime = time.Now().UTC()
	go svc.startHealthCheck()
	return resp
}

func (svc *service) startHealthCheck() {
	time.Sleep(initHealthCheckWait)
	for {
		if svc.recorder.state == idle {
			break
		}
		if svc.recorder.state != reloading && time.Now().UTC().Sub(svc.recorder.lastHealthCheckedAt) > maxUnhealthyRecorderInterval {
			go svc.reloadRecInChrome()
		}
		time.Sleep(healthCheckInterval)
	}
}

func (svc *service) reloadRecInChrome() {
	if err := setReloading(svc.recorder); err != nil {
		svc.logger.Log("err", errors.Wrap(err, "error while setting reloading state"))
		return
	}
	err := svc.recorder.ChromeCmd.Restart()
	if err != nil {
		svc.logger.Log("err", errors.Wrap(err, "error while restarting chrome due to bad health"))
		setRunning(svc.recorder)
		return
	}
	time.Sleep(initHealthCheckWait)
	if err := setRunningFromReloading(svc.recorder); err != nil {
		svc.logger.Log("err", errors.Wrap(err, "error while setting running state after reloading"))
	}
}

func (svc *service) Stop(ctx context.Context, req StopRecordingRequest) (resp StopRecordingResponse) {
	svc.recorder.mtx.Lock()
	defer svc.recorder.mtx.Unlock()

	if svc.recorder.state == idle {
		resp.Err = errors.New("avcapture: is not running", AlreadyEnded)
		return resp
	}
	setIdle(svc.recorder)
	stopTime := time.Now().UTC()
	err := errors.
		Do(svc.recorder.FFmpegCmd.Stop).
		Do(svc.recorder.ChromeCmd.Stop).
		Err()

	if err != nil {
		resp.Err = errors.New("avcapture: end running process error", err)
		return resp
	}
	cleanup(svc.recorder)
	resp.StopTime = stopTime
	return resp
}

func (svc *service) Check(context.Context) error {
	svc.recorder.lastHealthCheckedAt = time.Now().UTC()
	return nil
}
