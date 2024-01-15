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
COLLECTOR_CONTAINER_NAME?=cnf-collector
COLLECTOR_NS?=cnf-collector
GRAFANA_CONTAINER_NAME?=grafana
COLLECTOR_VERSION?=v0.0.3
REGISTRY?=quay.io
HOST_PORT?=80
TARGET_PORT?=80

COMMON_GO_ARGS=-race
GIT_COMMIT=$(shell scripts/create-version-files.sh)
GIT_RELEASE=$(shell scripts/get-git-release.sh)
GIT_PREVIOUS_RELEASE=$(shell scripts/get-git-previous-release.sh)
GOLANGCI_VERSION=v1.55.1
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
	docker pull ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION} 

stop-running-collector-container:
	docker ps -q --filter "name=${COLLECTOR_CONTAINER_NAME}" | xargs -r docker stop
	docker ps -aq --filter "name=${COLLECTOR_CONTAINER_NAME}" | xargs -r docker rm

# Runs collector locally with docker
run-collector: clone-tnf-secrets stop-running-collector-container
	docker run -d -p ${HOST_PORT}:${TARGET_PORT} --name ${COLLECTOR_CONTAINER_NAME} \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL='localhost' \
		-e DB_PORT='3306' \
		-e SERVER_ADDR=':${HOST_PORT}' \
		-e SERVER_READ_TIMEOUT=10 \
		-e SERVER_WRITE_TIMEOUT=10 \
		-e AWS_ACCESS_KEY=$(shell jq -r ".CollectorAWSAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e AWS_SECRET_ACCESS_KEY=$(shell jq -r ".CollectorAWSSecretAccessKey" "./tnf-secrets/collector-secrets.json") \
		${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION}
	rm -rf tnf-secrets

# Runs collector on rds with docker
run-collector-rds: clone-tnf-secrets stop-running-collector-container
	docker run -d -p ${HOST_PORT}:${TARGET_PORT} --name ${COLLECTOR_CONTAINER_NAME} \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL='collector-db.cn9luyhgvfkp.us-east-1.rds.amazonaws.com' \
		-e DB_PORT='3306' \
		-e SERVER_ADDR=':${HOST_PORT}' \
		-e SERVER_READ_TIMEOUT=10 \
		-e SERVER_WRITE_TIMEOUT=10 \
		-e AWS_ACCESS_KEY=$(shell jq -r ".CollectorAWSAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e AWS_SECRET_ACCESS_KEY=$(shell jq -r ".CollectorAWSSecretAccessKey" "./tnf-secrets/collector-secrets.json") \
		${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION}
	rm -rf tnf-secrets

# Runs collector on rds with docker in headless mode
run-collector-rds-headless: clone-tnf-secrets stop-running-collector-container
	docker run -d --name ${COLLECTOR_CONTAINER_NAME} -p ${HOST_PORT}:${TARGET_PORT} \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL='collector-db.cn9luyhgvfkp.us-east-1.rds.amazonaws.com' \
		-e DB_PORT='3306'\
		-e SERVER_ADDR=':${HOST_PORT}' \
		-e SERVER_READ_TIMEOUT=10 \
		-e SERVER_WRITE_TIMEOUT=10 \
		-e AWS_ACCESS_KEY=$(shell jq -r ".CollectorAWSAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e AWS_SECRET_ACCESS_KEY=$(shell jq -r ".CollectorAWSSecretAccessKey" "./tnf-secrets/collector-secrets.json") \
		-d ${COLLECTOR_IMAGE_NAME}
	rm -rf tnf-secrets

# Builds collector image locally
build-image-collector:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG} \
		-f Dockerfile .

# Builds collector image with latest and version tags
build-image-collector-by-version:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG} \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION} \
		-f Dockerfile .

# Pushes collector image with latest tag
push-image-collector:
	docker push ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG}

# Pushes collector image with latest tag and version tags
push-image-collector-by-version:
	docker push ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG}
	docker push ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION}

run-initial-mysql-scripts: clone-tnf-secrets
	sed \
		-e 's|CollectorAdminUser|$(shell jq -r ".CollectorAdminUser" "./tnf-secrets/collector-secrets.json" | base64 -d)|g' \
		-e 's|CollectorAdminPassword|$(shell jq -r ".CollectorAdminPassword" "./tnf-secrets/collector-secrets.json")|g' \
		./scripts/database/create_schema.sql > create_schema_prod.sql
	sed \
		-e 's/MysqlUsername/$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)/g' \
		-e 's/MysqlPassword/$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)/g' \
		./scripts/database/create_user.sql > create_user_prod.sql
	mysql -uroot -p < create_schema_prod.sql	# enter local mysql root password
	mysql -uroot -p < create_user_prod.sql		# enter local mysql root password
	rm create_schema_prod.sql create_user_prod.sql
	rm -rf tnf-secrets


# Deploys collector for CI test purposes
deploy-collector-for-CI:
	oc apply -f ${COLLECTOR_DEPLOYMENT_PATH} -n ${COLLECTOR_NS}

# Deploys mysql for CI test purposes
deploy-mysql-for-CI:
	oc apply -f ${MYSQL_DEPLOYMENT_PATH} -n ${COLLECTOR_NS}

stop-running-grafana-container:
	docker ps -q --filter "name=${GRAFANA_CONTAINER_NAME}" | xargs -r docker stop
	docker ps -aq --filter "name=${GRAFANA_CONTAINER_NAME}" | xargs -r docker rm

run-grafana: clone-tnf-secrets stop-running-grafana-container
	sed \
		-e 's/MysqlUsername/$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)/g' \
		-e 's/MysqlPassword/$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)/g' \
		./grafana/datasource/datasource-config.yaml > datasource-config-prod.yaml
	docker run -d -p 3000:3000 --name=${GRAFANA_CONTAINER_NAME} \
	  	-e "GF_SECURITY_ADMIN_USER=$(shell jq -r ".GrafanaUsername" "./tnf-secrets/collector-secrets.json")" \
  		-e "GF_SECURITY_ADMIN_PASSWORD=$(shell jq -r ".GrafanaPassword" "./tnf-secrets/collector-secrets.json")" \
		-v ./grafana/dashboard:/etc/grafana/provisioning/dashboards \
		-v ./datasource-config-prod.yaml:/etc/grafana/provisioning/datasources/datasource-config-prod.yaml \
		grafana/grafana
	rm datasource-config-prod.yaml
	rm -rf tnf-secrets

# Clones tnf-secret private repo if does not exist
clone-tnf-secrets:
	rm -rf tnf-secrets
	git clone git@github.com:test-network-function/tnf-secrets.git 
