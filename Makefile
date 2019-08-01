CONTAINER_IMAGE=etherlabsio/avcapture
CONTAINER_TAG=staging2

build:
	@docker build -t ${CONTAINER_IMAGE}:${CONTAINER_TAG} .
	@docker push ${CONTAINER_IMAGE}:${CONTAINER_TAG}
