#!/bin/bash

IMAGE="budget2-web-prod"
REGISTRY="trow.kube-public:31000"

go build main.go

docker build . -t ${IMAGE}:latest
docker image tag ${IMAGE}:latest ${REGISTRY}/${IMAGE}:latest
docker push ${REGISTRY}/${IMAGE}:latest

kubectl -n budget2 apply -f k8s/01_b2web
