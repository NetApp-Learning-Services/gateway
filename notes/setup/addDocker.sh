#!/bin/bash

sudo apt update
sudo apt install docker.io
sudo usermod -aG docker $USER
newgrp docker