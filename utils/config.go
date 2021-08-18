package config

import (
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/rest"
)

var Config *rest.Config

func init() {
	if env.IN_CLUSTER {
		var err error
		Config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		var err error
		Config, err = clientcmd.BuildConfigFromFlags("", env.KUBECONFIG)
		if err != nil {
			panic(err.Error())
		}
	}
}
