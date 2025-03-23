## setup kind for local development

```bash
kind create cluster --name faas
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
kubectl create token -n kubernetes-dashboard default
```

## Setup Harbor

```bash
helm repo add harbor https://helm.goharbor.io
helm repo update
helm install harbor harbor/harbor -f harbor-values.yml
```

## Setup knative

```bash

kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-core.yaml
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.17.0/kourier.yaml

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

kubectl --namespace kourier-system get service kourier

kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-default-domain.yaml

```

port forward the kourier service

```bash
kubectl port-forward svc/net-kourier-controller -n kourier-system 80:80
```

## Test Kservice

```bash
kubectl apply -f kservice.yml
```

## Setup ingress

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx

kubectl port-forward svc/nginx-ingress-ingress-nginx-controller 5000:80
```
