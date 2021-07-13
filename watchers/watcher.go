package watchers

import (
	"github.com/beopencloud/network-watcher/utils"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

var stopper chan struct{}

func Watch(restConfig *rest.Config) {
	client, err := kubernetes.NewForConfig(restConfig)
	exitOnErr(err)
	dynamic, err := dynamic.NewForConfig(restConfig)
	exitOnErr(err)
	k8sClient := utils.ExtendedClient{CoreClient: client, DynamicClient: dynamic}
	if stopper != nil {
		close(stopper)
	}
	stopper = make(chan struct{})
	go serviceWatch(k8sClient, stopper)
	go ingressWatch(k8sClient, stopper)

}

func exitOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
