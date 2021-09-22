# networkWhatcher
## Pre requis
* Go
    
      wget https://golang.org/dl/go1.16.6.linux-amd64.tar.gz
      rm -rf /usr/local/go && tar -C /usr/local -xzf go1.16.6.linux-amd64.tar.gz
      export PATH=$PATH:/usr/local/go/bin
      go version 

## Run on local

    make run
## Login to your registry image
    docker login -u username -p password
## Build and Push image
    make docker-build docker-push
## Deploy on cluster
Open the file Makefile and on the line 52 ( Make : manifests kustomize ), remove **manifests kustomize**

    make deploy


After deploying, two pods will be created
* one for API
* one for operator

Just the namespace that have the label **intrabpce.fr/network-watching=true** will be able to trace the logs 
     
     kubectl label ns default intrabpce.fr/network-watching=true
