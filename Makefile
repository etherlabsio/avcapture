#########################
# Environment Variables #	
#########################

SLACK_WEBHOOK_URL="https://hooks.slack.com/services/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"

# AWS details
AWS_ACCOUNT_ID=$(shell aws --profile ${ENV} sts get-caller-identity --output text --query 'Account')
AWS_ACCESS_KEY_ID=$(shell aws configure get aws_access_key_id --profile ${ENV})
AWS_SECRET_ACCESS_KEY=$(shell aws configure get aws_secret_access_key --profile ${ENV})
AWS_REGION=$(shell aws configure get region --profile ${ENV})

BRANCH=$(shell git rev-parse --short HEAD || echo -e '$CI_COMMIT_SHA')
DOCKER_LOGIN=$(shell aws --profile ${ENV} ecr get-login --no-include-email --region ${AWS_REGION})
IMAGE_PREFIX=${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/etherlabs
CONTAINER_IMAGE=${IMAGE_PREFIX}/avcapture

docker_login:
	@eval ${DOCKER_LOGIN}

build: docker_login
	@docker build--build-arg BASE_IMAGE=${CONTAINER_IMAGE} -t ${CONTAINER_IMAGE}:${BRANCH} .
	@docker push ${CONTAINER_IMAGE}:${BRANCH}
ifeq (${ENV},production)
	@docker tag ${CONTAINER_IMAGE}:${BRANCH} ${CONTAINER_IMAGE}:latest
	@docker push ${CONTAINER_IMAGE}:latest
else
	@docker tag ${CONTAINER_IMAGE}:${BRANCH} ${CONTAINER_IMAGE}:${ENV}
	@docker push ${CONTAINER_IMAGE}:${ENV}
endif

pre-deploy-notify:
	@curl -X POST --data-urlencode 'payload={"text": "[${ENV}] [${BRANCH}] ${USER}: avcapture is being deployed"}' ${SLACK_WEBHOOK_URL}

post-deploy-notify:
	@curl -X POST --data-urlencode 'payload={"text": "[${ENV}] [${BRANCH}] ${USER}: avcapture is deployed"}' ${SLACK_WEBHOOK_URL}

dev:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/capture github.com/etherlabsio/avcapture/cmd/capture
	env GOOS=linux GOARCH=amd64 go build -o ./bin/server github.com/etherlabsio/avcapture/cmd/server
	docker run --privileged --rm -p 3070:3070 --name avcapture -e ACTIVE_ENV=development -e FFMPEG_TGZ_URI="https://s3.amazonaws.com/io.etherlabs.artifacts/shared/ffmpeg-4.0.2-ubuntu-16.04.tgz" -e FFMPEG_DEPS_URI="https://s3.amazonaws.com/io.etherlabs.artifacts/shared/ffmpeg-libs-ubuntu-16.04.tgz" -v /tmp/avcapture:/data -v ${PWD}/scripts:/app -v ${PWD}/bin:/binaries  -it etherlabsio/avcapture 
