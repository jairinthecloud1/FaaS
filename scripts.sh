# build the image
docker build -t jairjosafath/faas-api .

# setup Harbor following the steps: https://gdservices.io/local-container-registry-with-harbor-and-minikube

# Create devops namespace
kubectl create namespace harbor

helm install harbor harbor/harbor --namespace harbor -f harbor-values.yml