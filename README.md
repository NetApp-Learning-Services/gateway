# Project Gateway 
A simple Kubernetes operator that creates, configures, and deletes NetApp ONTAP Storage Virtual Machine (SVM), also know as virtual server (vServer), resources.

## Description
This operator uses Red Hat's [Operator-SDK](https://sdk.operatorframework.io) to scaffold a controller that manages an Storage Virtual Machines (SVMs) resources in an NetApp ONTAP cluster. 


The current version of the operator is v1alpha2.

It currently creates and updates:
* an SVM, 
* an optional SVM management LIF, 
* an optional SVM administrator management credentials (vsadmin), 
* an optional NFS configuration with NFS interfaces and NFS exports, 
* and an optional iSCSI configuration with iSCSI interfaces.


When the custom resource (CR) is delete, the operator uses a finalizer (called gateway.netapp.com) to delete the SVM and all it configuration when the CR is deleted.  

## Getting Started

### 1. Install a version of the operator: 


```
kubectl create -f https://raw.githubusercontent.com/NetApp-Learning-Services/gateway/main/config/deploy/v1alpha2/gatewayoperator.yaml
```

### 2. Create a secret for the ONTAP cluster administrator's credentials:

(NOTE: This example is deployed in the gateway namesspaces that gets created when deploying the operator)

	
```
apiVersion: v1
kind: Secret
metadata:
  name: ontap-cluster-admin
  namespace: gateway
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: Netapp1!
```
	

### 3. Create an optional secret for the SVM administrator's credentials: 

(NOTE: This example is deployed in the gateway namesspaces that gets created when deploying the operator)
```
apiVersion: v1
kind: Secret
metadata:
  name: ontap-svm-admin
  namespace: gateway
type: kubernetes.io/basic-auth
stringData:
  username: vsadmin
  password: Netapp1!
```

### 4. Create a custom resource with your SVM settings:


(NOTE: Make sure you provide the required Cluster administrator's credentials created in step 2 and ```clusterHost``` with the NetApp ONTAP cluster management LIF. Also, the ```debug``` setting in the spec provides additional logging information in the operator's ```manager``` container logs. ) 

	
```
apiVersion: gateway.netapp.com/v1alpha2
kind: StorageVirtualMachine
metadata:
  name: storagevirtualmachine-testcase
  namespace: gateway-system
spec:
  svmName: testVs
  clusterHost: 192.168.0.102
  debug: false
  aggregates:
  - name: Cluster2_01_FC_1
  management:
    name: manage1
    ip: 192.168.0.30
    netmask: 255.255.255.0
    broadcastDomain: Default
    homeNode: Cluster2-01
  vsadminCredentials:
    name: ontap-svm-admin
    namespace: gateway-system 
  clusterCredentials:
    name: ontap-cluster-admin
    namespace: gateway-system
  iscsi:
    enabled: true
    alias: testVs
    interfaces:
    - name: iscsi1
      ip: 192.168.0.51
      netmask: 255.255.255.0
      broadcastDomain: Default
      homeNode: Cluster2-01
  nfs:
    enabled: true
    v3: true
    v4: true
    v41: true
    interfaces:
    - name: nfs1
      ip: 192.168.0.31
      netmask: 255.255.255.0
      broadcastDomain: Default
      homeNode: Cluster2-01
    export:
      name: default
      rules:
      - clients: 0.0.0.0/0
        protocols: any
        rw: any
        ro: any
        superuser: any
        anon:  "65534"
``` 

### 5. Deploy NetApp [Astra Trident](https://github.com/NetApp/trident) to manage the SVM resources created by this operator.

## Contributing
Written by Curtis Burchett

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster. 

## License

Copyright 2022.

Creative Commons Legal Code, CC0 1.0 Universal

