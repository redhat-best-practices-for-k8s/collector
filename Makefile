GO_PACKAGES=$(shell go list ./... | grep -v vendor)

# Get default value of $GOBIN if not explicitly set
GO_PATH=$(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
  GOBIN=${GO_PATH}/bin
else
  GOBIN=$(shell go env GOBIN)
endif

MYSQL_CONTAINER_NAME?=mysql-container
COLLECTOR_IMAGE_NAME?=testnetworkfunction/collector
COLLECTOR_IMAGE_TAG?=latest
COLLECTOR_CONTAINER_NAME?=tnf-collector
COLLECTOR_VERSION?=0.0.1
REGISTRY?=quay.io

COMMON_GO_ARGS=-race
GIT_COMMIT=$(shell scripts/create-version-files.sh)
GIT_RELEASE=$(shell scripts/get-git-release.sh)
GIT_PREVIOUS_RELEASE=$(shell scripts/get-git-previous-release.sh)
GOLANGCI_VERSION=v1.53.3
LINKER_TNF_RELEASE_FLAGS=-X github.com/test-network-function/cnf-certification-test/cnf-certification-test.GitCommit=${GIT_COMMIT}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/test-network-function/cnf-certification-test/cnf-certification-test.GitRelease=${GIT_RELEASE}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/test-network-function/cnf-certification-test/cnf-certification-test.GitPreviousRelease=${GIT_PREVIOUS_RELEASE}
MYSQL_DEPLOYMENT_PATH = ./k8s/deployment/database.yaml
COLLECTOR_DEPLOYMENT_PATH = ./k8s/deployment/app.yml

.PHONY: all clean test

# Build and run unit tests
test:
	go build ${COMMON_GO_ARGS} ./...
	UNIT_TEST="true" go test -coverprofile=cover.out.tmp ./...

vet:
	go vet ${GO_PACKAGES}

build:
	go build -ldflags "${LINKER_TNF_RELEASE_FLAGS}" ${COMMON_GO_ARGS} -o collector

# Installs linters
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_VERSION}

# Runs configured linters
lint:
	checkmake Makefile
	golangci-lint run --timeout 10m0s
	hadolint Dockerfile

install-mac-brew-tools:
	brew install \
		checkmake \
		golangci-lint \
		hadolint

# Pulls collector image from quay.io
pull-image-collector:
	docker pull ${COLLECTOR_IMAGE_NAME}

# Runs collector with docker
run-collector:
	docker run --network=host -p 8080:8080 --name ${COLLECTOR_CONTAINER_NAME} ${COLLECTOR_IMAGE_NAME}

# Runs collector with docker in headless mode
run-collector-headless:
	docker run --network=host --name ${COLLECTOR_CONTAINER_NAME} -p 8080:8080 -d ${COLLECTOR_IMAGE_NAME}

# Stops collector container
stop-collector:
	docker stop ${COLLECTOR_CONTAINER_NAME}

# Builds collector image locally
build-image-local:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG} \
		-f Dockerfile .

# Builds collector image with latest and version tags
build-image-collector-latest:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG} \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION} \
		-f Dockerfile .

# Deploy collector based on latest tag
deploy-collector-latest:
	oc apply -f ${COLLECTOR_DEPLOYMENT_PATH}

# Delete collector based on latest tag
delete-collector-latest:
	oc delete -f ${COLLECTOR_DEPLOYMENT_PATH}

# Builds collector image with dev tag
build-image-collector:
	docker build -f Dockerfile -t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:dev .

# Pushes collector image with dev tag
push-image-collector:
	docker push ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:dev

# Deploys collector based on dev tag
deploy-collector: clone-tnf-secrets
	sed \
		-e 's/latest/dev/g' \
		-e 's/MysqlUsername/$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json")/g' \
		-e 's/MysqlPassword/$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json")/g' \
		${COLLECTOR_DEPLOYMENT_PATH} > collector-deployment-dev.yml
	oc apply -f ./collector-deployment-dev.yml -n tnf-collector
	rm collector-deployment-dev.yml
	rm -rf ./tnf-secrets

# Removes collector image and deployment
delete-collector:
	docker rmi ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:dev
	oc delete -f ${COLLECTOR_DEPLOYMENT_PATH}

deploy-mysql: clone-tnf-secrets
	sed \
		-e "s/DBRootPassword/$(shell jq -r ".DBRootPassword" "./tnf-secrets/collector-secrets.json")/" \
		${MYSQL_DEPLOYMENT_PATH} > mysql-deployment-dev.yaml
	oc apply -f mysql-deployment-dev.yaml -n tnf-collector
	rm mysql-deployment-dev.yaml
	rm -rf ./tnf-secrets

delete-mysql:
	oc delete -f ${MYSQL_DEPLOYMENT_PATH}

clone-tnf-secrets:
	# temporarly using my fork's branch
	git clone -b collector-secrets git@github.com:shirmoran/tnf-secrets.git