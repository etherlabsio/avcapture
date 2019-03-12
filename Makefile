FFMPEG_IMAGE=etherlabsio/ffmpeg-avcapture
FFMPEG_TAG=latest
CONTAINER_IMAGE=etherlabsio/avcapture
CONTAINER_TAG=latest

build:
	@docker build -t ${FFMPEG_IMAGE}:${FFMPEG_TAG} -f ffmpeg-docker/Dockerfile .
	@docker push ${FFMPEG_IMAGE}:${FFMPEG_TAG}
	@docker build -t ${CONTAINER_IMAGE}:${CONTAINER_TAG} .
	@docker push ${CONTAINER_IMAGE}:${CONTAINER_TAG}
