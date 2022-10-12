# Project Astra Gateway 
A simple operator that creates, configures, and deletes ONTAP resources for Kubernetes

## Description
This operator uses Operator-SDK to scaffold a controller that used to managed Storage Virtual Machines (SVMs) resources in an ONTAP cluster.   It currently creates an SVM, the SVM management LIF when a custom resource (CR) is created and uses a finalizer (called gateway.netapp.com) to delete the SVM when the CR is deleted.  

## Getting Started
To get started, launch the Using Astra Control with Kuberentes Lab-on-Demand image.  

### Running on the cluster
1. Setup Kubernetes

```sh
Launch the Using Astra Control with Kuberentes Lab-on-Demand image.  
```

2. Setup a private docker registry
	
```sh
Install the private registry and configure all the nodes to use the have access to it.
```
	
3. Install Opeartor-SDK on one of the worker nodes

```sh
export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
export OS=$(uname | awk '{print tolower($0)}')
export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.23.0
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk
```

4. Install Go on the worker node
	
```sh
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.2.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
wget https://dl.google.com/go/go1.19.2.linux-amd64.tar.gz
sudo tar -C /usr/local/ -xzf go1.19.2.linux-amd64.tar.gz
```

5. Clone the repo on the worker node
	
```sh
git clone https://github.com/NetApp-Learning-Services/gateway/
```

6. Setup kubeconfig on the worker node
	
```sh
Copy kubeconfig from the master node if needed
```

7. Chnage directory into the repo location on the worker node and run the bash script
	
```sh
bash gateway\test.sh
```

## Contributing
Written by Curtis Burchett

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

## License

Copyright 2022.

Creative Commons Legal Code, CC0 1.0 Universal

