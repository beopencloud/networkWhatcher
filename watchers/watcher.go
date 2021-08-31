package watchers

import (
	"github.com/beopencloud/network-watcher/utils"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

var stopper chan struct{}

// +
// cette fonction est le point d'entrée du package watcher.
// c'est ici q'on initialise le client_go de kubernetes qu'on passe en parametre au watchers service et ingress
// +
func Watch(restConfig *rest.Config) {
	client, err := kubernetes.NewForConfig(restConfig)
	exitOnErr(err)
	dynamic, err := dynamic.NewForConfig(restConfig)
	exitOnErr(err)
	k8sClient := utils.ExtendedClient{CoreClient: client, DynamicClient: dynamic}
	if stopper != nil {
		close(stopper)
	}
	serviceWatch(k8sClient, stopper)

}

func exitOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
