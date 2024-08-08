# Cnf Certification's Collector

[![release)](https://img.shields.io/github/v/release/redhat-best-practices-for-k8s/collector?color=blue&label=%20&logo=semver&logoColor=white&style=flat)](https://github.com/redhat-best-practices-for-k8s/collector/releases)
[![red hat](https://img.shields.io/badge/red%20hat---?color=gray&logo=redhat&logoColor=red&style=flat)](https://www.redhat.com)
[![openshift](https://img.shields.io/badge/openshift---?color=gray&logo=redhatopenshift&logoColor=red&style=flat)](https://www.redhat.com/en/technologies/cloud-computing/openshift)

## Description

A Go-based endpoint for collecting
[Cnf Certification Suites](https://github.com/redhat-best-practices-for-k8s/certsuite)
results.

The CNF Certification Suites provide a set of test cases for the
Containerized Network Functions/Cloud Native Functions (CNFs) to verify if
best practices for deployment on Red Hat OpenShift clusters are followed.

The CNF Certification Suites results are saved in a `claim.json` file,
which in turn, could be sent to the Collector for storing its data.

The Collector can collect data by partner name or anonymously
(not saved under any partner name).\
Collecting data by your partner name,
will allow you to also get all the data saved under this name.

**Note 1:** Data saved anonymously won't be reachable
by the partner sending the data.\
**Note 2:** Collector's data collection is disabled by default,
and has to be enabled manually by the user running the CNF Certification Suites.

## How to use Collector?

### Send data to Collector

#### Option 1 - Run CNF Certification test suites and send results to Collector

1. Enable Collector's data collection:

    ```sh
    export TNF_ENABLE_DATA_COLLECTION=true
    ```

2. Adjust CNF Certification configuration file:\
    (see [CNF Certification configuration description](https://redhat-best-practices-for-k8s.github.io/certsuite/configuration/))

    1. Specify who executed the suites:\
    Under the `executedBy` entry in the `tnf_config.yml` file,
    specify who executed the CNF Certification suites (QE\CI\Partner).

    2. (Optional) Save your data in the Collector under partner name:
        * Option 1: Partner's first use of the Collector\
        Under the `partnerName` entry enter your partner name,
        and under the `collectorAppPassword` entry enter a password as you like.\
        **Note:** Make sure to save the password for future use.

        * Option 2: Partner who already used the Collector\
        Under the `partnerName` entry enter your partner name,
        and under the `collectorAppPassword` entry enter the password
        you defined in your first use of the collector.

    3. (Optional) Send your data to your own collector:\
        Under the `collectorAppEndpoint` entry enter your collector app
        endpoint.\
        **Note:** If won't be specified, the collector app endpoint
        will be set to CNF Certification's Collector app endpoint
        by default: <!-- markdownlint-disable -->
        http://claims-collector.cnf-certifications.sysdeseng.com
        <!-- markdownlint-enable -->

    Example of filled entries in CNF Certification configuration file,
    to allow data collection by partner name:

    ```sh
    executedBy: "Partner"
    partnerName: "partner_example"
    collectorAppPassword: "password_example"
    collectorAppEndpoint: "endpoint_example"
    ```

3. Run CNF Certification suites with the adjusted configuration file.\
 (see [CNF Certification Test description](https://redhat-best-practices-for-k8s.github.io/certsuite/test-container/))

#### Option 2 - send a claim.json file directly to Collector

If you haven't already, clone Collector's repository:

```sh
git pull https://github.com/redhat-best-practices-for-k8s/collector.git
```

From collector's repo root directory, use the following command:

<!-- markdownlint-disable -->
```sh
./scripts/send-to-collector.sh "enter_endpoint" "path/to/claim.json" "enter_executed_by" "enter_partner_name(optional)" "enter_password(optional)"
```
<!-- markdownlint-enable -->

<!-- markdownlint-disable -->
(CNF Certification's Collector app endpoint:
http://claims-collector.cnf-certifications.sysdeseng.com)
<!-- markdownlint-enable -->

### Get data from Collector

#### Option 1 - For both Admin and Partners

Partners who use collector to store data by their name,
can have access to their saved data.\
If you haven't already, clone Collector's repository:

```sh
git pull https://github.com/redhat-best-practices-for-k8s/collector.git
```

From collector's repo root directory, use the following command:

```sh
./scripts/get-from-collector.sh "enter_endpoint" "enter_partner_name" "enter_password"
```

<!-- markdownlint-disable -->
(CNF Certification's Collector app endpoint:
http://claims-collector.cnf-certifications.sysdeseng.com)
<!-- markdownlint-enable -->

See an output example:

```sh
[
        {
                "Claim": {
                        "id": 180788,
                        "cnf_version": "n/a, (non-OpenShift cluster)",
                        "executed_by": "ci",
                        "upload_time": "2024-03-20 11:49:33",
                        "partner_name": "ciuser_8357965459",
                        "s3_file_url": "ci/ciuser_8357965459/claim_2024-03-20-11:49:33"
                },
                "ClaimResults": [
                        {
                                "id": 15909169,
                                "claim_id": 180788,
                                "suite_name": "affiliated-certification",
                                "test_id": "affiliated-certification-operator-is-certified",
                                "test_status": "passed"
                        },
                        {
                                "id": 15909170,
                                "claim_id": 180788,
                                "suite_name": "lifecycle",
                                "test_id": "lifecycle-container-poststart",
                                "test_status": "passed"
                        },
                        ...
                ]
        }
]
```

#### Option 2 - For Admin only

Access the data through
[Collector's grafana dashboard](http://44.195.143.94:3000/d/e5530a23-24b9-4e7f-ab28-8e778d99f429/collector-s-dashboard?orgId=1).

## Run Collector Locally

### Prerequisites

* Docker or Podman
* MySQL

### Build and Run Collector's container locally

Use the following commands to build and run Collector's container and database locally:

* **Clone Collector's repository:**

    ```sh
    git pull https://github.com/redhat-best-practices-for-k8s/collector.git
    ```

* **(Optional) Build and Push your Collector image:**

    You can build your own collector image

    ```sh
    export REGISTRY="enter_your_registry"
    export COLLECTOR_IMAGE_NAME="enter_your_collector_image_name"
    export COLLECTOR_IMAGE_TAG="enter_your_collector_image_tag"
    make build-image-collector
    make push-image-collector
    ```

    **Note:** If skipping this step, the colletor container will use
    `quay.io/redhat-best-practices-for-k8s/collector:latest` image by default.

* **Initialize local Collector DB:**

    ```sh
    make run-initial-mysql-scripts
    ```

* **Run Collector's application via container:**

    ```sh
    export LOCAL_DB_URL=enter_your_local_IP_address
    make run-collector
    ```

* **Test it out:**

    1. Send data to your collector in one of the ways mentioned [above](https://github.com/redhat-best-practices-for-k8s/collector?tab=readme-ov-file#send-data-to-collector),
    setting the endpoint of your local collector app endpoint.

    2. Get the data from your collector using the
    [above instructions](https://github.com/redhat-best-practices-for-k8s/collector?tab=readme-ov-file#option-1---for-both-admin-and-partners)
    , setting the endpoint of your local collector app endpoint and
    credentials of the sent data.

* **Cleanup after:**

    ```sh
    make stop-running-collector-container
    ```
