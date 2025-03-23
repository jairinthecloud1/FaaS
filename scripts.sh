# build the image
docker build -t jairjosafath/faas-api .
docker push jairjosafath/faas-api
kubectl apply -f ./infra.yml
kubectl rollout restart deployment faas-api

# setup Harbor following the steps: https://gdservices.io/local-container-registry-with-harbor-and-minikube


helm upgrade harbor harbor/harbor -f harbor-values.yml
helm install harbor harbor/harbor -f harbor-values.yml


#add knative serving



kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-core.yaml
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.17.0/kourier.yaml

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

kubectl --namespace kourier-system get service kourier

kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-default-domain.yaml

 kubectl port-forward svc/kourier -n kourier-system 80:80

# add kind
cat <<EOF | kind create cluster -n faas --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx

kubectl port-forward svc/nginx-ingress-ingress-nginx-controller 80:80