package utils

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
	"github.com/beopencloud/network-watcher/k8s"
)


var Config *rest.Config

func init() {
	if k8s.IN_CLUSTER {
		var err error
		Config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		var err error
		Config, err = clientcmd.BuildConfigFromFlags("", k8s.KUBECONFIG)
		if err != nil {
			panic(err.Error())
		}
	}
}