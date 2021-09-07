package utils

import (
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
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

// ++
// +
// Cette fonction permet de verifier si un namespace est monitoré ou pas par l'operator.
// Pour qu'un namespace soit monitoré par l'operator, il faut qu'il ait le label beopenit.com/network-watching=true.
// Si le namespace n'a pas le label, les events create,update, delete service|ingress seront ignoré.
// +
// ++

func CheckNamespaceAutoGen(k8sClient ExtendedClient, namespaceName string) (bool, error) {
	namespace, err := k8sClient.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		log.Println("Error when getting namespace: ", err.Error())
		return false, err
	}
	var watch = false
	for key, value := range namespace.Labels {
		if key == "CnocdNamespaceLabelKey" && value == "CnocdNamespaceLabelValue" {
			watch = true
			break
		}
	}
	return watch, nil
}

/*
====== Scenario===========
*/
func SetLoabBalancerIP(k8sClient ExtendedClient, service *corev1.Service, ip string) error {
	service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: ip}}
	_, err := k8sClient.CoreV1().Services(service.Namespace).UpdateStatus(context.TODO(), service, metav1.UpdateOptions{})
	fmt.Println("ERROR", err)
	return err
}

func GetNamespaceIP(k8sClient ExtendedClient, namespaceName string) (string, error) {
	namespace, err := k8sClient.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		log.Println("Error when getting namespace: ", err.Error())
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
	var err error
	if service.Spec.Type == "ClusterIP" {
		err = k8sClient.CoreV1().Services(service.Namespace).Delete(context.TODO(), service.Name, metav1.DeleteOptions{})
		fmt.Println("Error", err)
	}

	return err
}

// for type clusterIP or NodePort
func PatchFakeServiceToSetIP(k8sClient ExtendedClient, service *corev1.Service, ip string) (name string, error error) {
	service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: ip}}
	serviceName := service.Name
	_, err := k8sClient.CoreV1().Services(service.Namespace).UpdateStatus(context.TODO(), service, metav1.UpdateOptions{})
	return serviceName, err
}

// for type clusterIP or NodePort
func PatchFakeServiceToDeleteIP(k8sClient ExtendedClient, service *corev1.Service, ip string) error {
	service.Spec.Type = "NodePort"
	_, err := k8sClient.CoreV1().Services(service.Namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	return err
}

func GetAllServices(k8sClient ExtendedClient, namespace string) (*corev1.ServiceList, error) {
	listService, err := k8sClient.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	return listService, err
}

func CreateFakeService(k8sClient ExtendedClient, service *corev1.Service) error {
	_, err := k8sClient.CoreV1().Services(service.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	return err
}
