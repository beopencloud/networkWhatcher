package utils

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"log"
)

type CoreClient kubernetes.Interface
type DynamicClient dynamic.Interface

type ExtendedClient struct {
	CoreClient
	DynamicClient
}

func CheckNamespaceAutoGen(k8sClient ExtendedClient, namespaceName string) (bool, error) {
	namespace, err := k8sClient.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		log.Println("Error when getting namespace: ", err.Error())
		return false, err
	}
	var watch = false
	for key, value := range namespace.Labels {
		if key == CnocdNamespaceLabelKey && value == CnocdNamespaceLabelValue {
			watch = true
			break
		}
	}
	return watch, nil
}
