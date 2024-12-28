# Project Gateway 
A simple Kubernetes operator that creates, configures, and deletes NetApp ONTAP Storage Virtual Machine (SVM), also know as virtual server (vServer), resources.

## Description
This operator uses Red Hat's [Operator-SDK](https://sdk.operatorframework.io) to scaffold a controller that manages an Storage Virtual Machines (SVMs) resources in an NetApp ONTAP cluster. 

The current version of the operator is v1beta2.  V1beta2 migrated Operator-SDK to 1.34.1 and updated go dependencies to avoid critical warnings.  

The operator creates and updates:
* an SVM, 
* an optional SVM management LIF, 
* an optional SVM administrator management credentials (vsadmin), 
* an optional NFS configuration with NFS interfaces and NFS exports, 
* an optional iSCSI configuration with iSCSI interfaces,
* an optional NVMe/TCP configure with NVMe/TCP interfaces,
* and an optional S3 configuration with S3 interfaces, users, HTTP/HTTPS, and buckets.

When the custom resource (CR) is delete, the operator uses a finalizer (called gateway.netapp.com/finalizer) to delete the SVM and all it configuration when the CR is deleted. NOTE: You will loose SVM's data when the CR is deleted.   

## Getting Started

### 1. Install a version of the operator: 

```
kubectl create -f https://raw.githubusercontent.com/NetApp-Learning-Services/gateway/main/config/deploy/v1beta2/gatewayoperator.yaml
```

### 2. Create a secret for the ONTAP cluster administrator's credentials:

(NOTE: This example is deployed in the gateway-system namesspaces that gets created when deploying the operator)

```
apiVersion: v1
kind: Secret
metadata:
  name: ontap-cluster-admin
  namespace: gateway-system
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: Netapp1!
```
	
### 3. Create an optional secret for the SVM administrator's credentials: 

(NOTE: This example is deployed in the gateway-system namespaces that gets created when deploying the operator)
```
apiVersion: v1
kind: Secret
metadata:
  name: ontap-svm-admin
  namespace: gateway-system
type: kubernetes.io/basic-auth
stringData:
  username: vsadmin
  password: Netapp1!
```

### 4. Create a custom resource with your SVM settings:

(NOTE: Make sure you provide the required Cluster administrator's credentials created in step 2 and ```clusterHost``` with the NetApp ONTAP cluster management LIF. Also, the ```debug``` setting in the spec provides additional logging information in the operator's ```manager``` container logs. ) 
	
```
apiVersion: gateway.netapp.com/v1beta2
kind: StorageVirtualMachine
metadata:
  name: svmsrc
  namespace: gateway-system
spec:
  svmName: svmsrc
  svmDeletionPolicy: Delete
  clusterHost: 192.168.0.101
  debug: true
  aggregates:
  - name: Cluster1_01_FC_1
  - name: Cluster1_01_FC_2
  management:
    name: manage1
    ip: 192.168.0.30
    netmask: 255.255.255.0
    broadcastDomain: Default
    homeNode: Cluster1-01
  vsadminCredentials:
    name: ontap-svmsrc-admin
    namespace: gateway-system
  clusterCredentials:
    name: ontap-cluster1-admin
    namespace: gateway-system
  s3:
    enabled: true
    name: svmsrc
    http:
      enabled: true
      port: 80
    https:
      enabled: true
      port: 443
      caCertificate:
        commonName: svmsrc-ca
        type: root-ca
        expiryTime: P725DT
    users:
    - name: gateway-s3-src
      namespace: gateway-system
    interfaces:
    - name: s31
      ip: 192.168.0.34   
      netmask: 255.255.255.0
      broadcastDomain: Default
      homeNode: Cluster1-01
    buckets:
    - name: tp-src
      size: 102005473280
      type: s3
  iscsi:
    enabled: true
    alias: svmsrc
    interfaces:
    - name: iscsi1
      ip: 192.168.0.32
      netmask: 255.255.255.0
      broadcastDomain: Default
      homeNode: Cluster1-01
  nvme:
    enabled: true
    interfaces:
    - name: nvme1
      ip: 192.168.0.33
      netmask: 255.255.255.0
      broadcastDomain: Default
      homeNode: Cluster1-01
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
      homeNode: Cluster1-01
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

#### Deletion Policy
The svmDeletionPolicy can be either Delete or Retain (default).  If set to Delete, upon deletion of the CR, the SVM is deleted.  The default behavior (svmDeleteionPolicy set to Retain) is upon deletion of the CR, the SVM is not deleted but must be manually managed. 

#### S3
The S3 protocol needs either HTTP or HTTPS configured, at least one user, and a S3-enabled LIF.  If you enable HTTPS, you must provide the a common name of CA certificate.  If the CA cert for the SVM does not exist, the operator will create a self-signed CA (root-ca) certificate. The operator will then create a Certificate Signing Request (CSR) with the common name the same as the SVM name and then sign the CSR with the CA certificate.  Finally, the signed CSR will then be installed as a server certificate with SVM.  This enables HTTPS' TSL for the S3 server. For a command-line equilvant to these steps, see this [doc](https://docs.netapp.com/us-en/ontap/s3-config/create-install-ca-certificate-svm-task.html). Finally, create at least 1 bucket with a minimum size of 102005473280 bytes (95 GiB).

### 5. Deploy NetApp [Trident](https://github.com/NetApp/trident) to manage the SVM resources created by this operator.

## Contributing
Written by Curtis Burchett

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster. 

## License
Copyright 2025.

Creative Commons Legal Code, CC0 1.0 Universal