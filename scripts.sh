# build the image
docker build -t jairjosafath/faas-api .
docker push jairjosafath/faas-api
kubectl apply -f .\infra.yml
kubectl rollout restart deployment faas-api

# setup Harbor following the steps: https://gdservices.io/local-container-registry-with-harbor-and-minikube


helm upgrade harbor harbor/harbor -f harbor-values.yml
helm install harbor harbor/harbor -f harbor-values.yml

kubectl delete -f .\infra.yml

kubectl apply -f .\infra.yml

kubectl delete -f ./infra.yml

kubectl apply -f ./infra.yml


#add knative serving

kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-core.yaml
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.17.0/kourier.yaml

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

kubectl --namespace kourier-system get service kourier

