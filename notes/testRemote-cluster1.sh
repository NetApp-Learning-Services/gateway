#!/bin/bash
# A simple bash script to reset and test the application
kubectl config use-context source-admin@source
kubectl -n gateway-system delete svm svmtest
kubectl -n gateway-system delete secret ontap-cluster1-admin
kubectl -n gateway-system delete secret ontap-svmtest-admin
kubectl delete namespace gateway-system 
#git pull origin v1beta2 #need to update when changing feature branches

#Uncomment below to install for the first time
#sudo apt install sshpass
sshpass -p Netapp1! ssh root@192.168.0.61 "ctr -n k8s.io i rm docker.io/curtisab/gateway:v1beta2"
sshpass -p Netapp1! ssh root@192.168.0.62 "ctr -n k8s.io i rm docker.io/curtisab/gateway:v1beta2"
sshpass -p Netapp1! ssh root@192.168.0.63 "ctr -n k8s.io i rm docker.io/curtisab/gateway:v1beta2"
sshpass -p Netapp1! ssh root@192.168.0.64 "ctr -n k8s.io i rm docker.io/curtisab/gateway:v1beta2"

make docker-build docker-push
make deploy
# kubectl -n gateway-system create secret docker-registry myreg --docker-server=https://docker-registry:30001 --docker-username=admin --docker-password=Netapp1!
# kubectl -n gateway-system patch sa gateway-controller-manager -p '{\"imagePullSecrets\": [{\"name\": \"myreg\"}]}'
kubectl -n gateway-system create -f notes/testCR-cluster1.yaml
