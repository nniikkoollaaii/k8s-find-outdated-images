kubectl apply -f test/ns.yaml
kubectl apply -f test/deployment.yaml
docker-compose -f test/docker-compose.yaml up -d 