package utils

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"time"
)

type CoreClient kubernetes.Interface
type DynamicClient dynamic.Interface

type ExtendedClient struct {
	CoreClient
	DynamicClient
}

// ++
// +
// Cette fonction permet de verifier si un namespace est monitoré ou pas par l'operator.
// Pour qu'un namespace soit monitoré par l'operator, il faut qu'il ait le label intrabpce.fr/network-watching=true.
// Si le namespace n'a pas le label, les events create,update, delete service|ingress seront ignoré.
// +
// ++

func CheckNamespaceAutoGen(k8sClient ExtendedClient, namespaceName string) (bool, error) {
	namespace, err := k8sClient.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	var watch = false
	for key, value := range namespace.Labels {
		if key == NetworkWatcherNamespaceLabelKey && value == NetworkWatcherNamespaceLabelValue {
			watch = true
			break
		}
	}
	return watch, nil
}

func SetLoabBalancerIP(k8sClient ExtendedClient, service *corev1.Service, ip string) error {
	service.Spec.Type = "LoadBalancer"
	service.Spec.ExternalIPs = []string{ip}
	_, err := k8sClient.CoreV1().Services(service.Namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	return err
}

func GetNamespaceIP(k8sClient ExtendedClient, namespaceName string) (string, error) {
	namespace, err := k8sClient.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	for key, value := range namespace.Annotations {
		if key == "watching/namespaceIp" {
			return value, nil
		}
	}
	return "", errors.New("Namespace IP Not Found ")
}

func DeleteFakeService(k8sClient ExtendedClient, service *corev1.Service) error {
	err := k8sClient.CoreV1().Services(service.Namespace).Delete(context.TODO(), service.Name, metav1.DeleteOptions{})
	if err != nil && !k8sError.IsNotFound(err) {
		return err
	}
	listService, err := k8sClient.CoreV1().Services(service.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, v := range listService.Items {
		if v.Name == "fake-service" {
			time.Sleep(1 * time.Second)
			return DeleteFakeService(k8sClient, service)
		}
	}
	return nil
}

