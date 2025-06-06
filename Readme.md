# Faas

## Overview

FaaS is a serverless framework that allows you to deploy and manage functions as a service on Kubernetes. It provides a simple and efficient way to run your code in response to events, without the need to manage the underlying infrastructure.

For the C-4 model of this project please go to [C4 Model](assets/c4.svg)

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/jairinthecloud1/FaaS.git
cd FaaS
```

## Prerequisites

The following dependencies are required to run and develop with FaaS:

| Dependency   | Description                                              | Installation Link                                      |
|-------------|----------------------------------------------------------|--------------------------------------------------------|
| Docker      | Containerization platform                                | [Get Docker](https://docs.docker.com/get-docker/)      |
| Kubernetes  | Container orchestration system                           | [Kubernetes Tools](https://kubernetes.io/docs/tasks/tools/) |
| kubectl     | Kubernetes command-line tool                             | [kubectl](https://kubernetes.io/docs/tasks/tools/)     |
| Helm        | Kubernetes package manager                               | [Helm Install](https://helm.sh/docs/intro/install/)    |
| Kind        | Local Kubernetes clusters using Docker containers         | [Kind Quick Start](https://kind.sigs.k8s.io/docs/user/quick-start/) |
| Knative     | Kubernetes-based platform to deploy and manage serverless workloads | [Knative Install](https://knative.dev/docs/install/)   |
| Auth0       | Authentication and authorization as a service (with GitHub connection) | [Auth0 Quickstart](https://auth0.com/docs/quickstart) |

## Auth0 Setup

To enable authentication, you need to set up an Auth0 application with a GitHub connection:

1. Go to [Auth0 Dashboard](https://manage.auth0.com/).
2. Create a new Application (Regular Web Application).
3. In the application settings, set the callback URLs and allowed logout URLs as needed for your environment.
4. Go to **Authentication > Social** and enable the **GitHub** connection. Follow the prompts to configure your GitHub OAuth app if you haven't already.
5. In your Auth0 application, enable the GitHub connection under the Connections tab.
6. Copy the following values from your Auth0 dashboard:
   - AUTH0_CLIENT_ID
   - AUTH0_DOMAIN
   - AUTH0_CLIENT_SECRET
   - AUTH0_CALLBACK_URL
7. Base64 encode each value and use them as described in the Patch Auth0 secrets section below.

## setup kind for local development

```bash
kn quickstart kind --name faas
```

## Test Kservice

```bash
kubectl apply -f kservice.yml
kubectl get ksvc
```

## Setup dashboard for local development

```bash
helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
helm repo update

helm install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace -n kubernetes-dashboard
```

```bash
kubectl port-forward svc/kubernetes-dashboard-kong-proxy 8443:443 -n kubernetes-dashboard
```

## apply the resources

These include the service account, role and role binding for the dashboard

```bash
kubectl apply -f infra.yml
```

```bash
kubectl create token -n kubernetes-dashboard kubernetes-dashboard
```

## Setup ingress

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx
```

Port forward the ingress controller to access the services

```bash
kubectl port-forward svc/nginx-ingress-ingress-nginx-controller 8888:80

```

you will have to set the port in your requests during testing

## Finally patch the secrets the go application needs to communicate with Docker Hub

Replace `base64==` with the base64 encoded username and password for Docker Hub.

```bash
kubectl patch secret faas-api-secret -p '{"data":{"DOCKER_USERNAME":"base64==","DOCKER_PASSWORD":"base64="}}'
```

## Patch Auth0 secrets for the go application

Replace `base64==` with the base64 encoded values for each Auth0 variable from your `.env` file:

```bash
kubectl patch secret faas-api-secret -p '{"data":{"AUTH0_CLIENT_ID":"base64==","AUTH0_DOMAIN":"base64==","AUTH0_CLIENT_SECRET":"base64==","AUTH0_CALLBACK_URL":"base64=="}}'
```

```bash
kubectl patch secret faas-api-secret -p '{"data":{"COOKIE_SECRET":"base64=="}}'
```

## restart the faas-api deployment

```bash
kubectl rollout restart deployment/faas-api
```

## The auth flow is under construction

For now as a placeholder there is a middleware function that checks the headers.Authorization
per request for a value `Bearer valid-token`, this will be replaced with the actual implementation of
proper authN|Z

## Quick tests

set <www.faas.test> in your /etc/hosts file and run the following command

```bash
curl --location 'www.faas.test:8888/api/health'
```

```bash
curl --location 'www.faas.test:8888'
```
