package main

import (
	"net/http"
	"os"
	"time"

	"github.com/etherlabsio/pkg/logutil"
	"github.com/gin-gonic/gin"

	resty "gopkg.in/resty.v1"

	"github.com/etherlabsio/avcapture/internal/chrome"
	execute "github.com/etherlabsio/avcapture/internal/exec"
	"github.com/etherlabsio/avcapture/internal/ffmpeg"
	"github.com/etherlabsio/avcapture/pkg/recorder"
)

const (
	restyTimeOutSec = 5
	restyRetries    = 3
)

const (
	serviceName = "recorder"
)

func main() {
	logger := logutil.NewServerLogger(true)
	port := os.Getenv("AVCAPTURE_PORT")
	if "" == port {
		port = ":3080"
	}

	_, err := execute.ExecuteCommand(logger, "pulseaudio -D --exit-idle-time=-1")
	if err != nil {
		panic(err)
	}
	_, err = execute.ExecuteCommand(logger, "pacmd load-module module-virtual-sink sink_name=v1")
	if err != nil {
		panic(err)
	}
	_, err = execute.ExecuteCommand(logger, "pacmd set-default-sink v1")
	if err != nil {
		panic(err)
	}
	_, err = execute.ExecuteCommand(logger, "pacmd set-default-source v1.monitor")
	if err != nil {
		panic(err)
	}

	_, err = execute.ExecuteCommand(logger, "Xvfb :99 -screen 0 1280x720x16 &> xvfb.log &")
	if err != nil {
		panic(err)
	}

	var appCtxt appContext
	resty.SetTimeout(restyTimeOutSec * time.Second)
	resty.Retries(restyRetries)

	appCtxt.logger = logger
	appCtxt.isRunning = false
	appCtxt.ff = ffmpeg.NewBuilder()

	appCtxt.cr = chrome.NewBuilder()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.ErrorLogger())

	r.POST("/start_recording", func(c *gin.Context) {
		var req recorder.StartRecordingRequest
		var err error
		if err = c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"msg": "invalid start_recording request" + err.Error()})
		}
		if err = handleRequest(&req, &appCtxt, recordingStarted); err != nil {
			logger.Log("error", "started", "error-stack", err)
			c.JSON(http.StatusInternalServerError,
				gin.H{"msg": "start_recording request failed with error" + err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"start_time": appCtxt.recSTime})
		}
	})
	r.POST("/stop_recording", func(c *gin.Context) {
		var err error
		if err = handleRequest(nil, &appCtxt, recordingEnded); err != nil {
			logger.Log("error", "ended", "error-stack", err)
			c.JSON(http.StatusInternalServerError,
				gin.H{"msg": "stop_recording request failed with error" + err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"stop_time": appCtxt.recETime})
		}
	})

	r.GET("/debug/healthcheck", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.Run(port)

}
