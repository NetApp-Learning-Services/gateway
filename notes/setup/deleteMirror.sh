#!/bin/bash

sudo apt install sshpass
sshpass -p Netapp1! ssh root@192.168.0.61 "rm /etc/containerd/certs.d/docker.io/hosts.toml"
sshpass -p Netapp1! ssh root@192.168.0.62 "rm /etc/containerd/certs.d/docker.io/hosts.toml"
sshpass -p Netapp1! ssh root@192.168.0.63 "rm /etc/containerd/certs.d/docker.io/hosts.toml"
sshpass -p Netapp1! ssh root@192.168.0.64 "rm /etc/containerd/certs.d/docker.io/hosts.toml"
sshpass -p Netapp1! ssh root@192.168.0.96 "rm /etc/containerd/certs.d/docker.io/hosts.toml"
sshpass -p Netapp1! ssh root@192.168.0.97 "rm /etc/containerd/certs.d/docker.io/hosts.toml"
sshpass -p Netapp1! ssh root@192.168.0.98 "rm /etc/containerd/certs.d/docker.io/hosts.toml"
