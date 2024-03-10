sudo apt install sshpass
sshpass -p Netapp1! ssh root@192.168.0.61 "ctr -n k8s.io i rm docker.io/curtisab/gateway:v1beta1"
sshpass -p Netapp1! ssh root@192.168.0.62 "ctr -n k8s.io i rm docker.io/curtisab/gateway:v1beta1"
sshpass -p Netapp1! ssh root@192.168.0.63 "ctr -n k8s.io i rm docker.io/curtisab/gateway:v1beta1"
sshpass -p Netapp1! ssh root@192.168.0.64 "ctr -n k8s.io i rm docker.io/curtisab/gateway:v1beta1"