# build the image
docker build -t jairjosafath/faas-api .

# setup Harbor following the steps: https://gdservices.io/local-container-registry-with-harbor-and-minikube


helm install harbor harbor/harbor -f harbor-values.yml