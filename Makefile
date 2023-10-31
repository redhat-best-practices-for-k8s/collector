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
	docker pull ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION} 

# Runs collector locally with docker
run-collector: clone-tnf-secrets
	docker run --network=host -p 8080:8080 --name ${COLLECTOR_CONTAINER_NAME} \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL='localhost' \
		-e DB_PORT='3306' \
		${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION}
	rm -rf tnf-secrets

# Runs collector on rds with docker
run-collector-rds: clone-tnf-secrets
	docker run --network=host -p 8080:8080 --name ${COLLECTOR_CONTAINER_NAME} \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL='collector-db.cn9luyhgvfkp.us-east-1.rds.amazonaws.com' \
		-e DB_PORT='3306' \
		${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION}
	rm -rf tnf-secrets

# Runs collector on rds with docker in headless mode
run-collector-rds-headless: clone-tnf-secrets
	docker run --network=host --name ${COLLECTOR_CONTAINER_NAME} -p 8080:8080 \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL='collector-db.cn9luyhgvfkp.us-east-1.rds.amazonaws.com' \
		-e DB_PORT='3306'\
		-d ${COLLECTOR_IMAGE_NAME}
	rm -rf tnf-secrets

# Stops collector container
stop-collector:
	docker stop ${COLLECTOR_CONTAINER_NAME}

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
	oc apply -f ${COLLECTOR_DEPLOYMENT_PATH} -n tnf-collector

# Deploys mysql for CI test purposes
deploy-mysql-for-CI:
	oc apply -f ${MYSQL_DEPLOYMENT_PATH} -n tnf-collector

# Clones tnf-secret private repo (temprary Shir's fork and branch)
clone-tnf-secrets:
	git clone git@github.com:test-network-function/tnf-secrets.git
