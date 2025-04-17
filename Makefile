GO_PACKAGES=$(shell go list ./... | grep -v vendor)

# Get default value of $GOBIN if not explicitly set
GO_PATH=$(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
  GOBIN=${GO_PATH}/bin
else
  GOBIN=$(shell go env GOBIN)
endif

MYSQL_CONTAINER_NAME?=mysql-container
COLLECTOR_IMAGE_NAME?=redhat-best-practices-for-k8s/collector
COLLECTOR_IMAGE_NAME_LEGACY?=testnetworkfunction/collector
COLLECTOR_IMAGE_TAG?=latest
COLLECTOR_CONTAINER_NAME?=cnf-collector
COLLECTOR_NS?=cnf-collector
GRAFANA_CONTAINER_NAME?=grafana
COLLECTOR_VERSION?=latest
REGISTRY?=quay.io
HOST_PORT?=80
TARGET_PORT?=80
LOCAL_DB_URL?=localhost

COMMON_GO_ARGS=-race
GIT_COMMIT=$(shell scripts/create-version-files.sh)
GIT_RELEASE=$(shell scripts/get-git-release.sh)
GIT_PREVIOUS_RELEASE=$(shell scripts/get-git-previous-release.sh)
BASH_SCRIPTS=$(shell find . -name "*.sh" -not -path "./.git/*")
LINKER_TNF_RELEASE_FLAGS=-X github.com/redhat-best-practices-for-k8s/certsuite/certsuite.GitCommit=${GIT_COMMIT}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/redhat-best-practices-for-k8s/certsuite/certsuite.GitRelease=${GIT_RELEASE}
LINKER_TNF_RELEASE_FLAGS+= -X github.com/redhat-best-practices-for-k8s/certsuite/certsuite.GitPreviousRelease=${GIT_PREVIOUS_RELEASE}
MYSQL_DEPLOYMENT_PATH = ./k8s/deployment/database.yaml
COLLECTOR_DEPLOYMENT_PATH = ./k8s/deployment/app.yml
DB_URL = database-collectordb-1hykanj2mxdh.cn9luyhgvfkp.us-east-1.rds.amazonaws.com

S3_BUCKET_NAME?=cnf-suite-claims
S3_BUCKET_REGION?=us-east-1

.PHONY: all clean test

# Build and run unit tests
test:
	go build ${COMMON_GO_ARGS} ./...
	UNIT_TEST="true" go test -coverprofile=cover.out.tmp ./...

vet:
	go vet ${GO_PACKAGES}

build:
	go build -ldflags "${LINKER_TNF_RELEASE_FLAGS}" ${COMMON_GO_ARGS} -o collector

# Runs configured linters
lint:
	checkmake --config=.checkmake Makefile
	golangci-lint run --timeout 10m0s
	hadolint Dockerfile
	shfmt -d scripts/*.sh
	typos
	markdownlint '**/*.md'
	yamllint --no-warnings .
	shellcheck --format=gcc ${BASH_SCRIPTS}

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

# Runs collector locally with docker using latest tag
run-collector: clone-tnf-secrets stop-running-collector-container
	docker run -d --pull always -p ${HOST_PORT}:${TARGET_PORT} --name ${COLLECTOR_CONTAINER_NAME} \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL=${LOCAL_DB_URL} \
		-e DB_PORT='3306' \
		-e SERVER_ADDR=':${HOST_PORT}' \
		-e SERVER_READ_TIMEOUT=10 \
		-e SERVER_WRITE_TIMEOUT=10 \
		-e AWS_ACCESS_KEY=$(shell jq -r ".CollectorAWSAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e AWS_SECRET_ACCESS_KEY=$(shell jq -r ".CollectorAWSSecretAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e S3_BUCKET_NAME=${S3_BUCKET_NAME} \
		-e S3_BUCKET_REGION=${S3_BUCKET_REGION} \
		${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG}
	rm -rf tnf-secrets

# Runs collector on rds with docker
run-collector-rds: clone-tnf-secrets stop-running-collector-container
	docker run --restart always -d --pull always -p ${HOST_PORT}:${TARGET_PORT} --name ${COLLECTOR_CONTAINER_NAME} \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL='${DB_URL}' \
		-e DB_PORT='3306' \
		-e SERVER_ADDR=':${HOST_PORT}' \
		-e SERVER_READ_TIMEOUT=10 \
		-e SERVER_WRITE_TIMEOUT=10 \
		-e AWS_ACCESS_KEY=$(shell jq -r ".CollectorAWSAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e AWS_SECRET_ACCESS_KEY=$(shell jq -r ".CollectorAWSSecretAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e S3_BUCKET_NAME=${S3_BUCKET_NAME} \
		-e S3_BUCKET_REGION=${S3_BUCKET_REGION} \
		${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION}
	rm -rf tnf-secrets

# Runs collector on rds with docker in headless mode
run-collector-rds-headless: clone-tnf-secrets stop-running-collector-container
	docker run -d --pull always --name ${COLLECTOR_CONTAINER_NAME} -p ${HOST_PORT}:${TARGET_PORT} \
		-e DB_USER='$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_PASSWORD='$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)' \
		-e DB_URL='${DB_URL}' \
		-e DB_PORT='3306'\
		-e SERVER_ADDR=':${HOST_PORT}' \
		-e SERVER_READ_TIMEOUT=10 \
		-e SERVER_WRITE_TIMEOUT=10 \
		-e AWS_ACCESS_KEY=$(shell jq -r ".CollectorAWSAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e AWS_SECRET_ACCESS_KEY=$(shell jq -r ".CollectorAWSSecretAccessKey" "./tnf-secrets/collector-secrets.json") \
		-e S3_BUCKET_NAME=${S3_BUCKET_NAME} \
		-e S3_BUCKET_REGION=${S3_BUCKET_REGION} \
		-d ${COLLECTOR_IMAGE_NAME}
	rm -rf tnf-secrets

# Builds collector image with latest tag
build-image-collector:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG} \
		-f Dockerfile .

build-image-collector-legacy:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME_LEGACY}:${COLLECTOR_IMAGE_TAG} \
		-f Dockerfile .

# Builds collector image with latest and version tags
build-image-collector-by-version:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG} \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION} \
		-f Dockerfile .

build-image-collector-by-version-legacy:
	docker build \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME_LEGACY}:${COLLECTOR_IMAGE_TAG} \
		-t ${REGISTRY}/${COLLECTOR_IMAGE_NAME_LEGACY}:${COLLECTOR_VERSION} \
		-f Dockerfile .

# Pushes collector image with latest tag
push-image-collector:
	docker push ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG}

# Pushes collector image with latest tag and version tags
push-image-collector-by-version:
	docker push ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_IMAGE_TAG}
	docker push ${REGISTRY}/${COLLECTOR_IMAGE_NAME}:${COLLECTOR_VERSION}

create-initial-mysql-scripts: 
	sed \
		-e 's|CollectorAdminUser|$(shell jq -r ".CollectorAdminUser" "./tnf-secrets/collector-secrets.json" | base64 -d)|g' \
		-e 's|CollectorAdminPassword|$(shell jq -r ".CollectorAdminPassword" "./tnf-secrets/collector-secrets.json")|g' \
		./scripts/database/create_schema.sql > create_schema_prod.sql
	sed \
		-e 's/MysqlUsername/$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)/g' \
		-e 's/MysqlPassword/$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)/g' \
		./scripts/database/create_user.sql > create_user_prod.sql

# Runs initial mysql scripts locally
run-initial-mysql-scripts: clone-tnf-secrets create-initial-mysql-scripts
	mysql -uroot -p < create_schema_prod.sql	# enter local mysql root password
	mysql -uroot -p < create_user_prod.sql		# enter local mysql root password
	rm create_schema_prod.sql create_user_prod.sql
	rm -rf tnf-secrets

# Runs initial mysql scripts on RDS instance
run-initial-mysql-scripts-rds: clone-tnf-secrets create-initial-mysql-scripts
	mysql \
		-h ${DB_URL} \
		-u$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d) \
		-p$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)
		< create_schema_prod.sql
	mysql \
		-h ${DB_URL} \
		-u$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d) \
		-p$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d) \
		< create_user_prod.sql
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

run-grafana: clone-tnf-secrets clone-certsuite-overview stop-running-grafana-container

# Replace the collector's datasource config variables.
	sed \
		-e 's/MysqlUsername/$(shell jq -r ".MysqlUsername" "./tnf-secrets/collector-secrets.json" | base64 -d)/g' \
		-e 's/MysqlPassword/$(shell jq -r ".MysqlPassword" "./tnf-secrets/collector-secrets.json" | base64 -d)/g' \
		./grafana/datasource/datasource-config.yaml > datasource-config-prod.yaml \

# Replace the certsuite-overview's datasource config variables.
	sed \
		-e 's/DB_USER/$(shell jq -r ".MysqlUser" "./tnf-secrets/certsuite-overview-secrets.json")/g' \
		-e 's/DB_PASSWORD/$(shell jq -r ".MysqlPassword" "./tnf-secrets/certsuite-overview-secrets.json" | base64 -d)/g' \
		-e 's/DB_URL/$(shell jq -r ".MysqlURL" "./tnf-secrets/certsuite-overview-secrets.json")/g' \
		-e 's/DB_PORT/$(shell jq -r ".MysqlPort" "./tnf-secrets/certsuite-overview-secrets.json")/g' \
		./certsuite-overview/grafana/datasource/datasource.yaml > datasource-certsuite-overview.yaml \

# Copy the certsuite-overview's dashboards to the collector's dashboard folder
	cp ./certsuite-overview/grafana/dashboard/dashboard.json grafana/dashboard/dashboard-certsuite-overview.json \

	docker run -d -p 3000:3000 --name=${GRAFANA_CONTAINER_NAME} \
	  	-e "GF_SECURITY_ADMIN_USER=$(shell jq -r ".GrafanaUsername" "./tnf-secrets/collector-secrets.json")" \
  		-e "GF_SECURITY_ADMIN_PASSWORD=$(shell jq -r ".GrafanaPassword" "./tnf-secrets/collector-secrets.json")" \
		-v ./grafana/dashboard:/etc/grafana/provisioning/dashboards \
		-v ./datasource-config-prod.yaml:/etc/grafana/provisioning/datasources/datasource-config-prod.yaml \
		-v ./datasource-certsuite-overview.yaml:/etc/grafana/provisioning/datasources/datasource-certsuite-overview.yaml \
		grafana/grafana
	rm datasource-config-prod.yaml
	rm datasource-certsuite-overview.yaml
	rm -rf tnf-secrets

# Clones tnf-secret private repo if does not exist
clone-tnf-secrets:
	rm -rf tnf-secrets
	git clone git@github.com:redhat-best-practices-for-k8s/tnf-secrets.git

clone-certsuite-overview:
	rm -rf certsuite-overview
	git clone git@github.com:redhat-best-practices-for-k8s/certsuite-overview.git

