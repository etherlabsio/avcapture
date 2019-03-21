# avcapture

avcapture allows you to run a containarized application that captures the content and pipes the audio/video for any URL for encoding including generating a live playlist.

## Build and Run

- **Build**: `make build`
- **Run**: 

`docker run -it --net test -v $PWD/path/to/dir:$PWD/path/to/dir -v $PWD/path/to/data:/data -e FFMPEG_URI="<url for ffmpeg executable in tgz>" -e FFMPEG_DEPS_URI="<url for ffmpeg dependencies in lib directory>" --name avcapture -p 8080:8080 etherlabsio/avcapture`

  - As part of ini script, avcapture will download ffmpeg and dependencies to `/data` directory. ffmpeg binary will be copied to /usr/local/bin and dependencies will be copied to `/usr/local/lib`
  - If the ffmpeg and dependencies have to be extracted in every run, initialize the environment variable `DISABLE_FFMPEG_CACHE` with some value. For example `-e DISABLE_FFMPEG_CACHE="true"`.

## Dependencies

* A link to a tar.gz containing FFmpeg v4.0.2 binary to be provided by the end user as an environment variable on docker run.

## Configuration

* By default, the application will run server on port `8080`. If the application has to run on a different port, set the environment variable `PORT` with the new port inside the Dockerfile before the `ENTRYPOINT`.
* The user must provide a link to download their own distribution of `FFmpeg v4.0.2` via the `FFMPEG_TGZ_URI` environment variable during docker run.
* A URI containing `tar.gz` distribution of shared libs to be used by the FFmpeg binary can be provided setting the `FFMPEG_DEPS_URI` environment variable.

## API

### start-recording

- POST: <http://IP:8080/start_recording>
- "ffmpeg:options" and "chrome:options" are optional. If it is specified, the original arguments to these applications will be replaced completely with the provided one. User has to take care on the arguments passed for proper functionality.

```json
{
  "ffmpeg": {
    "params": [
      ["-hls_time", "6"],
      ["-hls_playlist_type", "event"],
      ["-hls_segment_filename", "/work/out%04d.ts"],
      ["/work/play.m3u8"]
    ],
    "options": [
      ["-y", ""],
      ["-v", "info"],
      ["-f", "x11grab"],
      ["-draw_mouse", "0"],
      ["-r", "24"],
      ["-s", "1280x720"],
      ["-thread_queue_size", "4096"],
      ["-i", ":99.0+0,0"],
      ["-f", "pulse"],
      ["-thread_queue_size", "4096"],
      ["-i", "default"],
      ["-acodec", "aac"],
      ["-strict", "-2"],
      ["-ar", "48000"],
      ["-c:v", "libx264"],
      ["-x264opts", "no-scenecut"],
      ["-preset", "veryfast"],
      ["-profile:v", "main"],
      ["-level", "3.1"],
      ["-pix_fmt", "yuv420p"],
      ["-r", "24"],
      ["-crf", "25"],
      ["-g", "48"],
      ["-keyint_min", "48"],
      ["-force_key_frames", "\"expr:gte(t,n_forced*2)\""],
      ["-tune", "zerolatency"],
      ["-b:v", "3600k"],
      ["-maxrate", "4000k"],
      ["-bufsize", "5200k"]
    ]
  },
  "chrome": {
    "url": "<https://www.youtube.com/watch?v=Bey4XXJAqS8>",
    "options": [
      "--enable-logging=stderr",
      "--autoplay-policy=no-user-gesture-required",
      "--no-sandbox",
      "--start-maximized",
      "--window-position=100,300",
      "--window-size=1280,720"
    ]
  }
}
```

### stop-recording

- POST: <http://IP:8080/stop_recording>
- No parameter is passed to this call.

## Output

User is supposed to map a directory from host system to the docker image. Along with this, user has to provide the output path (as part of `/start_recording` api) which will direct the output generated to the corresponding directory.

## Architecture

The docker image contains google chrome, ffmpeg and wrapper application.
As part of startup, the wrapper application will configure the system to run chrome browser on the given display id and to capture audio using pulseaudio.
Once the `/start_recording` is received, chrome will be started to render the `url` provided. An instance of ffmpeg will be started to capture the display.

## Notice  

We **do not** package avcapture with either FFmpeg or any shared libs and expect the end user to provide their own distribution.
