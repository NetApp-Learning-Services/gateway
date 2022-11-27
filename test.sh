#!/bin/bash
# A simple bash script to reset and test the application

kubectl -n gateway-system delete svm storagevirtualmachine-testcase
kubectl -n gateway-system delete secret ontap-cluster-admin
kubectl -n gateway-system delete secret ontap-svm-admin
kubectl delete namespace gateway-system 
git pull origin v1alpha2 #need to update when changing feature branches
make docker-build docker-push IMG=docker-registry:30001/curtisab/gateway:v1alpha2
make deploy
kubectl -n gateway-system create -f notes/testCR.yaml
