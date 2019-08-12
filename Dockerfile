FROM golang AS go-builder

WORKDIR $GOPATH/src/app

# Force the go compiler to use modules
ENV GO111MODULE=on

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

# This is the ‘magic’ step that will download all the dependencies that are specified in
# the go.mod and go.sum file.
# Because of how the layer caching system works in Docker, the  go mod download
# command will _ only_ be re-run when the go.mod or go.sum file change
# (or when we add another docker instruction this line)
RUN go mod download

# ADD . . blows up the build cache. Avoid using it when possible and predictable.
COPY cmd cmd
COPY internal internal
COPY pkg pkg

RUN CGO_ENABLED=0 go build -tags debug -o /dist/server -v -i -ldflags="-s -w" ./cmd/server

FROM etherlabsio/avcapture:base

WORKDIR /app
RUN apt-get update && \
    apt-get install -y graphicsmagick --no-install-recommends && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
COPY scripts/run-chrome.sh run-chrome.sh
COPY scripts/start-server.sh start-server.sh
RUN /bin/sh run-chrome.sh

ENV DISPLAY=:99
ENV LD_LIBRARY_PATH=/usr/local/lib

COPY --from=go-builder /dist /bin/

## Hack to remove default  browser check in chrome
ENTRYPOINT ["/app/start-server.sh"]
