#!/bin/bash

export GO_VERSION=1.23.6

export PATH=$PATH:/usr/local/go/bin
wget https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz
sudo tar -C /usr/local/ -xzf go$GO_VERSION.linux-amd64.tar.gz
export GO111MODULE=on