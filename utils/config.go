package utils

import (
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/rest"
)

var Config *rest.Config

func init() {
	if IN_CLUSTER {
		var err error
		Config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		var err error
		Config, err = clientcmd.BuildConfigFromFlags("", KUBECONFIG)
		if err != nil {
			panic(err.Error())
		}
	}
}
