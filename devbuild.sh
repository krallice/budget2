#!/bin/bash

IMAGE="budget2-web-dev"
REGISTRY="trow.kube-public:31000"

go build main.go

docker build . -t ${IMAGE}:latest
docker image tag ${IMAGE}:latest ${REGISTRY}/${IMAGE}:latest
docker push ${REGISTRY}/${IMAGE}:latest

kubectl -n budget2 apply -f k8s/11_b2webdev

sleep 5
POD=$(kubectl -n budget2 get pod | grep b2web-dev | head -n1 | awk '{print $1}')
kubectl -n budget2 logs -f $POD
