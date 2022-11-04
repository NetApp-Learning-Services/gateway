#!/bin/bash
# A simple bash script to upload to docker hub

git pull origin
make docker-build docker-push IMG=curtisab/gateway:v1alpha1
# make deploy
# kubectl -n gateway-system create -f notes/testCR.yaml
