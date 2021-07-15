# networkWhatcher
## Pre requis
* Go

## Run on local

    make run
  
## Deploy on cluster
Open the file Makefile and on the line 52 ( Make : manifests kustomize ), remove **manifests kustomize**

    make deploy


After deploying, two pods will be created
* one for API
* one for operator

Just the namespace that have the label **beopenit.com/network-watching=true** will be able to trace the logs 
     
     kubectl label ns default beopenit.com/network-watching=true
