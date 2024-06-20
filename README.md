# SPIFFE Demo

Repo that contains a SPIFFE Demo

## Structure

This repo contains the following:

* Golang tool
* Terraform
* Kubernetes deployment
* Dockerfiles

### Golang tool

A simple Golang tool to showcase SPIFFE possibilities. It has 3 subcommands:

1. customer
2. backend
3. httpbackend

The customer is the entry point for customers through an Ingress. It serves a simple webserver that is exposed over an Ingress and shows a page with buttons that allows an end-user to take actions. The following actions can be taken:

1. Connect to a SPIFFE server backend. This connects to another application that runs with the backend subcommand. The connection is SPIFFE authenticated and authorized. This showcases the potential when SPIFFE is integrated in the application layer.
1. Connect to a non-SPIFFE server backend. connects to another application that runs with the httpbackend subcommand. In the Kubernetes deployment we have put an Envoy in front that will authenticate and authorize the SPIFFE connection. This showcases the potential when SPIFFE can't be integrated in the application layer
1. Talk to AWS Services. This writes and reads from an AWS S3 bucket with a SPIFFE JWT identity. In the container we abstract everything away from the application (for the application it is as it would run natively in AWS). This is done through the spiffe-aws-assume-role binary.

### Terraform

The setup of the OIDC federation between our SPIRE install and AWS happens through this. It also creates the necessary S3 bucket and IAM roles and policies so our customer application can authenticate to AWS.

### Kubernetes deployment

The different manifests to deploy to Kubernetes. Currently this is opinionated to my setup.

### Dockerfiles

We need to package our Golang tool so it can be deployed in our Kubernetes cluster. For the customer application we also require an init-container that sets the AWS config correctly.

## Setup

### SPIRE

Install SPIRE by changing the values (the domain and trust domains) in `./deploy/spire/values.yaml` and installing it with Helm

```bash
helm upgrade --install -n spire spire-crds spire-crds --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace
helm upgrade --install -n spire spire spire -f ./deploy/spire/values.yaml --repo https://spiffe.github.io/helm-charts-hardened/
```
