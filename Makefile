CONTAINER_IMAGE=etherlabsio/avcapture
CONTAINER_TAG=latest

build:
	@docker build -t ffmpeg-ubuntu16.04:latest -f ffmpeg-docker/Dockerfile .
	@docker build -t ${CONTAINER_IMAGE}:${CONTAINER_TAG} .
	@docker push ${CONTAINER_IMAGE}:${CONTAINER_TAG}
