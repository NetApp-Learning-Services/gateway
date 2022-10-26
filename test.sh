#!/bin/bash
# A simple bash script to reset and test the application

kubectl -n gateway-system delete svm storagevirtualmachine-sample
kubectl -n gateway-system delete secret ontap-cluster-admin
kubectl -n gateway-system delete secret ontap-svm-admin
kubectl delete namespace gateway-system 
git pull origin
make docker-build docker-push IMG=docker-registry:30001/astra/gateway:v0.1
make deploy
kubectl -n gateway-system create -f notes/testCR.yaml
