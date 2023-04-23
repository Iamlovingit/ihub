docker build -t hfproxy:latest -d Dockerfile .
docker tag hfproxy:latest harbor-infp.com:14444/ais-system47/hfproxy:latest
docker push harbor-infp.com:14444/ais-system47/hfproxy:latest
kubectl delete -f service.yaml
kubectl apply -f service.yaml