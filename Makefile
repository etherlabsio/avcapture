CONTAINER_IMAGE=etherlabsio/avcapture
CONTAINER_TAG=staging

build:
	@docker build -t ${CONTAINER_IMAGE}:${CONTAINER_TAG} .
	@docker push ${CONTAINER_IMAGE}:${CONTAINER_TAG}
