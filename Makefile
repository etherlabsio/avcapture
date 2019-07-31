CONTAINER_IMAGE=etherlabsio/avcapture
CONTAINER_TAG=staging2

build:
	@docker build -t ${CONTAINER_IMAGE}:${CONTAINER_TAG} .
	@docker push ${CONTAINER_IMAGE}:${CONTAINER_TAG}

dev:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/capture github.com/etherlabsio/avcapture/cmd/capture
	env GOOS=linux GOARCH=amd64 go build -o ./bin/server github.com/etherlabsio/avcapture/cmd/server
	docker run --privileged --rm -p 3070:3070 --name avcapture -e ACTIVE_ENV=development -e FFMPEG_TGZ_URI="https://s3.amazonaws.com/io.etherlabs.artifacts/shared/ffmpeg-4.0.2-ubuntu-16.04.tgz" -e FFMPEG_DEPS_URI="https://s3.amazonaws.com/io.etherlabs.artifacts/shared/ffmpeg-libs-ubuntu-16.04.tgz" -v /tmp/avcapture:/data -v ${PWD}/bin:/bin -v ${PWD}/scripts:/app/scripts -it etherlabsio/avcapture
