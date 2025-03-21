# SPIFFE Demo

This SPIFFE demo repository aims to showcase the possibilities of different SPIFFE use-cases and requirements end-users might have in real production environments to integrate with their applications.

## Structure

This repo contains the following:

* Golang application
* Terraform
* Kubernetes deployment
* Dockerfiles

### Architecture

The architecture of the SPIFFE-demo applications is the following when deployed:

![SPIFFE Demo Architecture](img/SPIFFE-Demo-Architecture.png)

### Golang application

A simple Golang tool to showcase SPIFFE possibilities. It has 3 subcommands:

1. customer
2. backend
3. httpbackend

The customer is the entry point for customers through an Ingress. It serves a simple webserver that is exposed over an Ingress and shows a page with buttons that allows an end-user to take actions. The following actions can be taken:

1. Connect to a SPIFFE server backend. This connects to another application that runs with the backend subcommand. The connection is SPIFFE authenticated and authorized. This showcases the potential when SPIFFE is integrated in the application layer.
1. Connect to a non-SPIFFE server backend. connects to another application that runs with the httpbackend subcommand. In the Kubernetes deployment we have put an Envoy in front that will authenticate and authorize the SPIFFE connection. This showcases the potential when SPIFFE can't be integrated in the application layer
1. Talk to AWS S3 Service. This writes and reads from an AWS S3 bucket with a SPIFFE JWT identity. In the container we abstract everything away from the application (for the application it is as it would run natively in AWS). This is done through the [spiffe-aws-assume-role](https://github.com/MattiasGees/spiffe-aws-assume-role) binary. That binary gets called through the AWS Profile [`credential_process`](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-sourcing-external.html). Alternatively you can also use the X.509 authentication with [AWS IAM Roles Anywhere](https://docs.aws.amazon.com/rolesanywhere/latest/userguide/introduction.html) and the [aws-spiffe-workload-helper](https://github.com/spiffe/aws-spiffe-workload-helper).
1. Talk to Google Cloud Service. This writes and reads from an GCS bucket with a SPIFFE JWT identity. In the container we abstract everything away from the application (for the application it is as it would run natively in Google Cloud). This is done through the [spiffe-gcp-proxy](https://github.com/GoogleCloudPlatform/professional-services/tree/main/tools/spiffe-gcp-proxy) proxy. That proxy gets called when making a call to the internal metadata API of Google Cloud.
1. Talk to a PostgreSQL database with its SVID. It writes a randomly generated user to a database every time you click the button. With the retrieval function it will retrieve all previous generated users from the database. No username or password authentication is required. It uses the [SPIFFE-helper](https://github.com/spiffe/spiffe-helper/) to let PostgreSQL consume the SVID that it got issued. The SPIFFE-helper is responsible for writing it to an in-memory filesystem that is accessible by the PostgreSQL container and than reloads the PostgreSQL config to make sure that PostgreSQL is aware of the latest certificates. As PostgreSQL doesn't understand SPIFFE IDs, it does verification based on the CN on the X.509. By configuring SPIRE in such a way, it will create those extra entries for the application SVID and that way it can authenticate and authorize itself to PostgreSQL
1. A SPIFFE retriever endpoint `HOSTNAME/spifferetriever` to show the SVID details.

### Terraform

The setup of the OIDC federation between our SPIRE install with AWS and Google Cloud happens through Terraform. It also creates the necessary GCS, S3 buckets and IAM roles and policies so our customer application can authenticate to AWS and Google Cloud.

### Kubernetes deployment

The different manifests to deploy the different components to Kubernetes.

### Dockerfiles

We need to package our Golang tool so it can be deployed in our Kubernetes cluster. For the customer application we also require an init-container that sets the AWS config correctly. All of these images have been published to [Docker hub](https://hub.docker.com/repository/docker/mattiasgees).

#### initcontainer

The initcontainer is responsible for creating an AWS Profile config and making that available to the `customer` application.

#### postgresql

Adds 2 extra things to the [default](https://hub.docker.com/_/postgres) PostgreSQL image:

1. `init-user-db.sh`: Creates the necessary database, user and tables in the PostgreSQL during initialization
1. `set-pg-hba.sh`: Overwrites the default `pg_hba.conf` during initialization to force authentication and authorization with certificate instead of the normal username/password

#### spiffe-helper

Creates a Docker image for the latest version of the SPIFFE-helper. The SPIFFE-helper will than be used as a sidecar container next to the PostgreSQL container to make sure PostgreSQL always has the latest SVID loaded.

#### spiffe-gcp-proxy

Creates a Docker image for the latests version of the [spiffe-gcp-proxy](https://github.com/GoogleCloudPlatform/professional-services/tree/main/tools/spiffe-gcp-proxy). The SPIFFE GCP Proxy will than be used as a sidecar container next to our customer application to intercept calls to Google Cloud and add the necessary SPIFFE authentication credentials to it.

#### Go application

The Golang application gets built with golang through a built container. Afterwards `ca-certificates` and the `spiffe-aws-assume-role` binary get added to it as well.

## Setup

### Prerequisites

To be able to run this demo, there are a few prerequisites

* Ingress controller that is publicly exposed. This is required to make the OIDC endpoint available as well as the demo application.
* cert-manager installed and configured with a ClusterIssuer `letsencrypt-prod` to be able to request certificates for your OIDC endpoint and the demo application.
* DNS hostnames configured for OIDC and the demo application. These need to your Ingress controller.

### Cloud

The AWS and Google Cloud bits are optional. When you deploy the `spiffe-demo` application on your Kubernetes cluster, it will deploy everything but off-course if you haven't done the cloud provider setup those specific bits will not work.

#### Optional

For AWS, you can also decide to either use JWT (OIDC Federation) or X.509 (AWS IAM Anywhere) authentication. The default is AWS, If you wish to use X.509, change the value of the environment variable `AWS_AUTH` to `X509`.

### Prepare environment

Some values are going to be specific to your environment. We are going to prepare these now:

```bash
export DEMO_HOSTNAME=demo.yourdomain.com
export DEMO_ROGUE_HOSTNAME=demo-rogue.yourdomain.com
export OIDC_HOSTNAME=oidc.yourdomain.com
export AWS_BUCKET_NAME=myrandombucketname
export AWS_AUTH=JWT
export GCP_BUCKET_NAME=mygcpbucketname
export GCP_PROJECT_NAME=my-gcp-project-name
export GCP_PROJECT_NUMBER=11111111

sed -i '' "s/OIDC_HOSTNAME/$OIDC_HOSTNAME/g" deploy/spire/values.yaml
sed -i '' "s/OIDC_HOSTNAME/$OIDC_HOSTNAME/g; s/AWS_BUCKET_NAME/$AWS_BUCKET_NAME/g; s/JWT/$AWS_AUTH/g" deploy/terraform/aws/variables.tf
sed -i '' "s/OIDC_HOSTNAME/$OIDC_HOSTNAME/g; s/GCP_BUCKET_NAME/$GCP_BUCKET_NAME/g; s/GCP_PROJECT_NAME/$GCP_PROJECT_NAME/g" deploy/terraform/google/variables.tf
sed -i '' "s/AWS_BUCKET_NAME/$AWS_BUCKET_NAME/g; s/GCP_BUCKET_NAME/$GCP_BUCKET_NAME/g; s/GCP_PROJECT_NAME/$GCP_PROJECT_NAME/g; s/GCP_PROJECT_NUMBER/$GCP_PROJECT_NUMBER/g; s/DEMO_HOSTNAME/$DEMO_HOSTNAME/g; s/DEMO_ROGUE_HOSTNAME/$DEMO_ROGUE_HOSTNAME/g; s/AWS_AUTH/$AWS_AUTH/g" deploy/chart/spiffe-demo/values.yaml
```

### SPIRE

Install SPIRE with Helm. Make sure that the OIDC endpoint is publicly available as AWS will need to be able to hit that endpoint. You can take a look at the `values.yaml` that is used.

```bash
helm upgrade --install -n spire spire-crds spire-crds --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace
# Change ./deploy/spire/values.yaml to match your environment before running the next command
helm upgrade --install -n spire spire spire -f ./deploy/spire/values.yaml --repo https://spiffe.github.io/helm-charts-hardened/
```

### Terraform

#### AWS

Make sure you have access to an AWS account through the CLI, Terraform will use the same method to create the necessary resources. Take a look at `deploy/terraform/aws/variables.tf` to verify the expected environment specific values.

##### (optional) X509

If you have set the environemnt variable `AWS_AUTH` to `X509` earlier, you need to  create a `root.pem` file with the content of your root CA of your SPIRE server in `deploy/terraform/aws`. You can retrieve the root CA the following way:

```bash
# Create temp pod
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: spire-tools
  name: spire-tools
  namespace: default
spec:
  containers:
  - image: mattiasgees/spire-tools:latest
    imagePullPolicy: Always
    name: spire-tools
    resources: {}
    volumeMounts:
    - mountPath: /spiffe-workload-api
      name: spiffe-workload-api
      readOnly: true
  volumes:
  - csi:
      driver: csi.spiffe.io
      readOnly: true
    name: spiffe-workload-api
EOF

# Get root.pem
kubectl exec -it spire-tools -- sh -c 'spire-agent api fetch -socketPath /spiffe-workload-api/socket -output json | jq -r ".svids[0].bundle" | awk "BEGIN {print \"-----BEGIN CERTIFICATE-----\"} {print} END {print \"-----END CERTIFICATE-----\"}" | fold -w 64 > /tmp/root.pem && cat /tmp/root.pem' > deploy/terraform/aws/root.pem

# Delete temp pod 
kubectl delete pod spire-tools

```

##### Deploy

```bash
cd deploy/terraform/aws
# Change variables.tf to match your environment before running the next command
terraform init
terraform apply
terraform output -json | jq -r '@sh "export AWS_ROLE_ARN=\(.role_arn.value)\nexport AWS_TRUST_ANCHOR_ARN=\(.trust_anchor_arn.value)\nexport AWS_PROFILE_ARN=\(.profile_arn.value)"' >env.sh
source env.sh
```

#### Google Cloud

Make sure you have access to an Google Cloud account. Take a look at `deploy/terraform/google/variables.tf` to verify the expected environment specific values.

```bash
cd deploy/terraform/google
# Change variables.tf to match your environment before running the next command
terraform init
terraform apply
```

### Kubernetes

Take a look `deploy/chart/spiffe-demo/values.yaml` to verify it matches your environment and after that do a Helm install.

```bash
sed -i '' "s/AWS_ROLE_ARN/$AWS_ROLE_ARN/g; s/AWS_TRUST_ANCHOR_ARN/$AWS_TRUST_ANCHOR_ARN/g; s/AWS_PROFILE_ARN/$AWS_PROFILE_ARN/g" deploy/chart/spiffe-demo/values.yaml
helm upgrade --install -n spiffe-demo2 spife-demo ./deploy/chart/spiffe-demo --create-namespace
```

## Contributing

Contributions of new use-cases or improvements are more than welcome!
