# build the image
docker build -t jairjosafath/faas-api .
docker push jairjosafath/faas-api
kubectl apply -f .\infra.yml
kubectl rollout restart deployment faas-api

# setup Harbor following the steps: https://gdservices.io/local-container-registry-with-harbor-and-minikube


helm upgrade harbor harbor/harbor -f harbor-values.yml
helm install harbor harbor/harbor -f harbor-values.yml


kubectl apply -f .\infra.yml
kubectl delete -f .\infra.yml