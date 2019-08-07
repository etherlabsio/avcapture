package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/oklog/run"
	"github.com/peterbourgon/ff"

	"github.com/etherlabsio/avcapture/internal/recorder"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/etherlabsio/avcapture/pkg/commander"
	"github.com/etherlabsio/errors"
	"github.com/etherlabsio/healthcheck"
	"github.com/etherlabsio/pkg/httputil"
	"github.com/etherlabsio/pkg/logutil"
	"github.com/go-kit/kit/log/level"
)

// Configuration defines the settings required during the app initiation
type Configuration struct {
	Port  *string
	Debug *bool
}

func setupAVCaptureDevices() (err error) {
	if runtime.GOOS != "linux" {
		return nil
	}
	commands := []string{
		"pulseaudio -D --exit-idle-time=-1",
		"pacmd load-module module-virtual-sink sink_name=v1",
		"pacmd set-default-sink v1",
		"pacmd set-default-source v1.monitor",
		"Xvfb :99 -screen 0 1280x720x16 &> xvfb.log &",
	}
	for _, cmd := range commands {
		err = commander.Exec(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func httpStatusCodeFrom(err error) int {
	switch errors.KindOf(err) {
	case errors.Invalid:
		return http.StatusBadRequest
	case errors.Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}

func main() {
	fs := flag.NewFlagSet("avcapture", flag.ExitOnError)
	config := Configuration{
		Port:  fs.String("port", ":8080", "Port for the HTTP listener"),
		Debug: fs.Bool("debug", false, "Enable debug mode"),
	}
	ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix())
	logger := logutil.NewServerLogger(true)

	err := setupAVCaptureDevices()
	if err != nil {
		level.Error(logger).Log("msg", "av devices setup failure", "err", err)
		os.Exit(1)
	}

	var recorderService recorder.Service
	{
		recorderService = recorder.NewService(logger)
		recorderService = recorder.ValidationMiddleware(recorderService)
		recorderService = recorder.LoggingMiddleware(logger)(recorderService)
	}

	defer recorderService.Stop(context.Background(), recorder.StopRecordingRequest{})

	httpErrorEncoder := httputil.JSONErrorEncoder(httpStatusCodeFrom)
	httpJSONResponseEncoder := httputil.EncodeJSONResponse(httpErrorEncoder)

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	})

	r := chi.NewRouter()
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler)

	r.Post("/start_recording", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req recorder.StartRecordingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpJSONResponseEncoder(ctx, w, errors.Serializable(err))
			return
		}
		resp := recorderService.Start(ctx, req)
		httpJSONResponseEncoder(ctx, w, resp)
	})

	r.Post("/stop_recording", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		resp := recorderService.Stop(ctx, recorder.StopRecordingRequest{})
		httpJSONResponseEncoder(ctx, w, resp)
	})

	r.Get("/debug/healthcheck", healthcheck.HandlerFunc(healthcheck.WithChecker("recorder", recorderService)))

	var g run.Group
	{
		// The HTTP listener mounts the Go kit HTTP handler we created.
		//address := cfg.GetString("http.address")
		//ETHER Temp change
		httpListener, err := net.Listen("tcp", *config.Port)
		if err != nil {
			logger.Log("transport", "HTTP", "during", "listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "HTTP", "addr", config.Port)
			return http.Serve(httpListener, r)
		}, func(error) {
			httpListener.Close()
		})
	}
	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	logger.Log("exit", g.Run())
}
