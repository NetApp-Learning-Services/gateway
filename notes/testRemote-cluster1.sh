#!/bin/bash
# A simple bash script to reset and test the application
VERSION=v1beta3

kubectl config use-context source-admin@source
kubectl -n gateway-system delete svm svmsrc
# kubectl -n gateway-system delete secret ontap-cluster1-admin
# kubectl -n gateway-system delete secret ontap-svmtest-admin
# kubectl delete namespace gateway-system 
make undeploy
#git pull origin v1beta2 #need to update when changing feature branches

#Uncomment below to install for the first time
#sudo apt install sshpass
#sshpass -p Netapp1! ssh root@192.168.0.61 "ctr -n k8s.io i rm docker.io/curtisab/gateway:$VERSION"
echo "192.168.0.62: "
sshpass -p Netapp1! ssh -o StrictHostKeyChecking=no root@192.168.0.62 "ctr -n k8s.io i rm docker.io/curtisab/gateway:$VERSION"
echo "192.168.0.63 "
sshpass -p Netapp1! ssh -o StrictHostKeyChecking=no root@192.168.0.63 "ctr -n k8s.io i rm docker.io/curtisab/gateway:$VERSION"
echo "192.168.0.64 "
sshpass -p Netapp1! ssh -o StrictHostKeyChecking=no root@192.168.0.64 "ctr -n k8s.io i rm docker.io/curtisab/gateway:$VERSION"

make docker-build docker-push
make deploy
# kubectl -n gateway-system create secret docker-registry myreg --docker-server=https://docker-registry:30001 --docker-username=admin --docker-password=Netapp1!
# kubectl -n gateway-system patch sa gateway-controller-manager -p '{\"imagePullSecrets\": [{\"name\": \"myreg\"}]}'
kubectl -n gateway-system create -f notes/testCR-cluster1.yaml
