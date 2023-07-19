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
GIT_COMMIT=$(shell script/create-version-files.sh)
GIT_RELEASE=$(shell script/get-git-release.sh)
GIT_PREVIOUS_RELEASE=$(shell script/get-git-previous-release.sh)
GOLANGCI_VERSION=v1.53.3
LINKER_TNF_RELEASE_FLAGS=-X github.com/test-network-function/cnf-certification-test/cnf-certification-test.GitCommit=${GIT_COMMIT}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/test-network-function/cnf-certification-test/cnf-certification-test.GitRelease=${GIT_RELEASE}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/test-network-function/cnf-certification-test/cnf-certification-test.GitPreviousRelease=${GIT_PREVIOUS_RELEASE}
CREATE_SCHEMA_RAW_URL = "https://raw.githubusercontent.com/test-network-function/collector-deployment/main/database/create_schema.sql"
CREATE_MYSQL_USER_RAW_URL = "https://raw.githubusercontent.com/test-network-function/collector-deployment/main/database/create_user.sql"
COLLECTOR_DEPLOYMENT_RAW_URL = "https://raw.githubusercontent.com/test-network-function/collector-deployment/main/k8s/collector-deployment.yml"
MYSQL_PV_PATH = ./k8s/mysql-pv.yaml
MYSQL_DEPLOYMENT_PATH = ./k8s/mysql-deployment.yaml
COLLECTOR_DEPLOYMENT_PATH = ./k8s/collector-deployment.yml

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

# Builds a local container based on mysql image
build-mysql-container-local:
	docker run --name ${MYSQL_CONTAINER_NAME} -e MYSQL_ROOT_PASSWORD=pa55 \
		 -h 127.0.0.1 -p 3306:3306 -d mysql:latest
		 
# Builds local schema
build-schema:
	curl -sSL -o create_schema.sql ${CREATE_SCHEMA_RAW_URL}
	docker exec -i ${MYSQL_CONTAINER_NAME} mysql -u root -ppa55 < create_schema.sql
	rm -f create_schema.sql

# Builds a local mysql user for the above container
build-mysql-user-local:
	curl -sSL -o create_user.sql ${CREATE_MYSQL_USER_RAW_URL}
	docker exec -i ${MYSQL_CONTAINER_NAME} mysql -u root -ppa55 < create_user.sql
	rm -f create_user.sql

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

build-image-collector:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG} \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION} \
		-f Dockerfile .

build-and-deploy-image-collector-dev:
	docker build -f Dockerfile -t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:dev
	docker push ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:dev
	curl ${COLLECTOR_DEPLOYMENT_RAW_URL} | sed 's/latest/dev/g' > collector-deployment.yml
	oc apply -f ./collector-deployment.yml
	rm collector-deployment.yml

remove-image-collector-and-deployment-dev:
	docker rmi ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:dev
	oc delete deployment collector-deployment

deploy-mysql:
	oc apply -f ${MYSQL_PV_PATH}
	oc apply -f ${MYSQL_DEPLOYMENT_PATH}

delete-mysql:
	oc delete -f ${MYSQL_DEPLOYMENT_PATH}
	oc delete -f ${MYSQL_PV_PATH}

deploy-collector:
	oc apply -f ${COLLECTOR_DEPLOYMENT_PATH}

delete-collector:
	oc delete -f ${COLLECTOR_DEPLOYMENT_PATH}