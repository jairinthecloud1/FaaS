# Faas-Local-Development

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kubernetes](https://kubernetes.io/docs/tasks/tools/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [Knative](https://knative.dev/docs/install/)

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

kubectl port-forward svc/nginx-ingress-ingress-nginx-controller 80:80
```

## Finally patch the secrets the go application needs to communicate with Docker Hub

```bash
kubectl patch secret faas-api-secret -p '{"data":{"DOCKER_USERNAME":"base64==","DOCKER_PASSWORD":"base64="}}'
```

## Optional: restart the faas-api deployment

```bash
kubectl rollout restart deployment/faas-api
```
