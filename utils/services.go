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
	service.Spec.Type = "LoadBalancer"
	updatedService, err := k8sClient.CoreV1().Services(service.Namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println("111 Error 101", err)
	}
	updatedService.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: ip}}
	_, err = k8sClient.CoreV1().Services(updatedService.Namespace).UpdateStatus(context.TODO(), updatedService, metav1.UpdateOptions{})

	fmt.Println("111 Error 102", err)
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
	log.Println("111 delete fake service service.Name==", service.Name)
	err := k8sClient.CoreV1().Services(service.Namespace).Delete(context.TODO(), service.Name, metav1.DeleteOptions{})
	if err != nil{
		return err
	}
	listService, err := k8sClient.CoreV1().Services(service.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, v := range listService.Items {
		if v.Name == "fake-service" {
			log.Println("111 deleting fake-service in progress ...")
			time.Sleep(2*time.Second)
			return DeleteFakeService(k8sClient, service)
		}
	}
	log.Println("111 Fake-service successfully deleted ...")
	return nil
}

// for type clusterIP or NodePort
func PatchFakeServiceToSetIP(k8sClient ExtendedClient, service *corev1.Service, ip string) (name string, error error) {
	service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: ip}}
	serviceName := service.Name
	_, err := k8sClient.CoreV1().Services(service.Namespace).UpdateStatus(context.TODO(), service, metav1.UpdateOptions{})
	return serviceName, err
}

// for type clusterIP or NodePort
func PatchFakeServiceToDeleteIP(k8sClient ExtendedClient, service *corev1.Service) error {
	service.Spec.Type = "NodePort"
	_, err := k8sClient.CoreV1().Services(service.Namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		log.Println("Error when getting namespace: ", err)
		return err
	}
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

func PatchFakeService(k8sClient ExtendedClient, service *corev1.Service, ip string) (ser *corev1.Service, error error) {
	service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: ip}}
	servicePatched, err := k8sClient.CoreV1().Services(service.Namespace).UpdateStatus(context.TODO(), service, metav1.UpdateOptions{})
	fmt.Println("PTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT")
	fmt.Println("Error 102", err)
	return servicePatched, err
}
