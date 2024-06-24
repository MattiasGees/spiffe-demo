# SPIFFE Demo

This SPIFFE demo repository aims to showcase the possibilities of different SPIFFE use-cases and requirements end-users might have in real production environments.

## Structure

This repo contains the following:

* Golang application
* Terraform
* Kubernetes deployment
* Dockerfiles

### Golang application

A simple Golang tool to showcase SPIFFE possibilities. It has 3 subcommands:

1. customer
2. backend
3. httpbackend

The customer is the entry point for customers through an Ingress. It serves a simple webserver that is exposed over an Ingress and shows a page with buttons that allows an end-user to take actions. The following actions can be taken:

1. Connect to a SPIFFE server backend. This connects to another application that runs with the backend subcommand. The connection is SPIFFE authenticated and authorized. This showcases the potential when SPIFFE is integrated in the application layer.
1. Connect to a non-SPIFFE server backend. connects to another application that runs with the httpbackend subcommand. In the Kubernetes deployment we have put an Envoy in front that will authenticate and authorize the SPIFFE connection. This showcases the potential when SPIFFE can't be integrated in the application layer
1. Talk to AWS Services. This writes and reads from an AWS S3 bucket with a SPIFFE JWT identity. In the container we abstract everything away from the application (for the application it is as it would run natively in AWS). This is done through the [spiffe-aws-assume-role](https://github.com/MattiasGees/spiffe-aws-assume-role) binary. That binary gets called through the AWS Profile [`credential_process`](TODO).
1. Talk to a PostgreSQL database with its SVID. It writes a randomly generated user to a database every time you click the button. With the retrieval function it will retrieve all previous generated users from the database. No username or password authentication is required. It uses the [SPIFFE-helper](https://github.com/spiffe/spiffe-helper/) to let PostgreSQL consume the SVID that it got issued. The SPIFFE-helper is responsible for writing it to an in-memory filesystem that is accessible by the PostgreSQL container and than reloads the PostgreSQL config to make sure that PostgreSQL is aware of the latest certificates. As PostgreSQL doesn't understand SPIFFE IDs, it does verification based on the CN on the X.509. By configuring SPIRE in such a way, it will create those extra entries for the application SVID and that way it can authenticate and authorize itself to PostgreSQL
1. A SPIFFE retriever endpoint `HOSTNAME/spifferetriever` to show the SVID details.

### Terraform

The setup of the OIDC federation between our SPIRE install and AWS happens through this. It also creates the necessary S3 bucket and IAM roles and policies so our customer application can authenticate to AWS.

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

#### Go application

The Golang application gets built with golang through a built container. Afterwards `ca-certificates` and the `spiffe-aws-assume-role` binary get added to it as well.

## Setup

### SPIRE

Install SPIRE by changing the values (the domain and trust domains) in `./deploy/spire/values.yaml` and installing it with Helm. Make sure that the OIDC endpoint is publicly available as AWS will need to be able to hit that endpoint.

```bash
helm upgrade --install -n spire spire-crds spire-crds --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace
# Change ./deploy/spire/values.yaml to match your environment before running the next command
helm upgrade --install -n spire spire spire -f ./deploy/spire/values.yaml --repo https://spiffe.github.io/helm-charts-hardened/
```

### Terraform

Make sure you have access to an AWS account through the CLI, Terraform will use the same method to create the necessary resources. Before running terraform, make sure to change the variables in `deploy/terraform/variables.tf` to match your environment.

```bash
cd deploy/terraform
# Change variables.tf to match your environment before running the next command
terraform init
terraform apply
```

### Kubernetes

TODO
