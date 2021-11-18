package watchers

import (
	"context"
	"errors"
	"fmt"
	"github.com/beopencloud/network-watcher/utils"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"net/http"
	"time"

	// We need this import to load the GCP auth plugin which is required to authenticate against GKE clusters.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var serviceWatcherLogger = logf.Log.WithName("service_watcher")

// ++
// +
// Voici l'implémentation du watcher de service.
// vous pouvez voir dans le code les fonctions de callback appelées, dépendance des events add, update et delete service.
// Au niveau de chaque fonction de callback, on récupère le service concerné et on envoie l'event à l'API.
// par exemple, lors de la création d'un service, nous obtenons l'objet du service nouvellement créé qu'on
// envoie à l'API pour lui notifier la creation du nouveau service.
// +
// ++
func serviceWatch(k8sClient utils.ExtendedClient, stopper chan struct{}) {
	factory := informers.NewSharedInformerFactory(k8sClient, 0)
	informer := factory.Core().V1().Services().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			go func(obj interface{}) {
				service := obj.(*corev1.Service)
				reqLogger := serviceWatcherLogger.WithValues("service", service.Name, "namespace", service.Namespace)
				watch, err := utils.CheckNamespaceAutoGen(k8sClient, service.Namespace)
				if !watch || err != nil {
					return
				}

				ip, err := utils.GetNamespaceIP(k8sClient, service.Namespace)
				if err != nil {
					reqLogger.Error(err, "Error to get IP from Annotation")
					return
				}

				if service.Spec.Type == "NodePort" && service.Labels["servicetype"] == "LoadBalancer" {
					// TODO Get fake-service and Delete fake-service
					listService, err := k8sClient.CoreV1().Services(service.Namespace).List(context.TODO(), metav1.ListOptions{})
					if err != nil {
						reqLogger.Error(err, "Error to Get List of Services")
						return
					}
					var fakeService corev1.Service
					for _, v := range listService.Items {
						if v.Name == "fake-service" {
							fakeService = v
						}
					}
					err = utils.DeleteFakeService(k8sClient, &fakeService)
					if err != nil {
						reqLogger.Error(err, "Error to Delete Fake-service")
						return
					}
					// TODO Patch service type to LoadBalancer et IP annotation
					err = utils.SetLoabBalancerIP(k8sClient, service, ip)
					if err != nil {
						reqLogger.Error(err, "Error to Patch Type Service and IP")
						return
					}
				} else {
					fmt.Println("No service to patched")
				}

				res, err := utils.PostRequestToAPI(service)
				if err != nil {
					reqLogger.Error(err, "Error to send service create event to API")
					return
				}
				if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
					responseBody, _ := ioutil.ReadAll(res.Body)
					reqLogger.Error(errors.New("Error to send service create event to API"), string(responseBody), "StatusCode", res.StatusCode)
					return
				}
				reqLogger.Info("service created")
			}(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			go func(oldObj, newObj interface{}) {
				service := newObj.(*corev1.Service)
				reqLogger := serviceWatcherLogger.WithValues("service", service.Name, "namespace", service.Namespace)
				watch, err := utils.CheckNamespaceAutoGen(k8sClient, service.Namespace)
				if !watch || err != nil {
					return
				}

				ip, err := utils.GetNamespaceIP(k8sClient, service.Namespace)
				if err != nil {
					reqLogger.Error(err, "Error to get IP from Annotation")
					return
				}
				err = utils.SetLoabBalancerIP(k8sClient, service, ip)
				if err != nil {
					reqLogger.Error(err, "Error to Patch Type Service and IP ")
					return
				}

				res, err := utils.PutRequestToAPI(service)
				if err != nil {
					reqLogger.Error(err, "Error to send service update event to API")
					return
				}
				if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
					responseBody, _ := ioutil.ReadAll(res.Body)
					reqLogger.Error(errors.New("Error to send service update event to API"), string(responseBody), "StatusCode", res.StatusCode)
					return
				}
				reqLogger.Info("service updated")

			}(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			go func(obj interface{}) {
				service := obj.(*corev1.Service)
				reqLogger := serviceWatcherLogger.WithValues("service", service.Name, "namespace", service.Namespace)
				watch, err := utils.CheckNamespaceAutoGen(k8sClient, service.Namespace)
				if !watch || err != nil {
					return
				}
				ip, err := utils.GetNamespaceIP(k8sClient, service.Namespace)
				if err != nil {
					reqLogger.Error(err, "Error to get IP from Annotation")
					return
				}

			getServices:
				services, err := k8sClient.CoreV1().Services(service.Namespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					reqLogger.Error(err, "Error to Get List of Services")
					return
				}
				for _, v := range services.Items {
					if v.Name == service.Name {
						time.Sleep(2 * time.Second)
						goto getServices
					}
				}
				recreateFakeService := true
				for _, v := range services.Items {
					if v.Name != "fake-service" && v.Spec.Type == "NodePort" {
						recreateFakeService = false
					}
				}
				if recreateFakeService {
					serviceAdd := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "fake-service",
							Namespace: service.Namespace,
						},
						Spec: corev1.ServiceSpec{
							Type: "LoadBalancer",
							Ports: []corev1.ServicePort{
								{
									Port: 80,
									Name: "http",
								},
								{
									Port: 443,
									Name: "https",
								},
							},
						},
					}
					newFakeService, err := k8sClient.CoreV1().Services(service.Namespace).Create(context.TODO(), serviceAdd, metav1.CreateOptions{})
					if err != nil {
						reqLogger.Error(err, "Error to create fake-service")
						return
					}
					_, err = utils.PatchFakeServiceToSetIP(k8sClient, newFakeService, ip)
					if err != nil {
						reqLogger.Error(err, "Error to patch service")
						return
					}

				}
				res, err := utils.DeleteRequestToAPI(service)
				if err != nil {
					reqLogger.Error(err, "Error to send service delete event to API")
					return
				}
				if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
					responseBody, _ := ioutil.ReadAll(res.Body)
					reqLogger.Error(errors.New("Error to send service delete event to API"), string(responseBody), "StatusCode", res.StatusCode)
					return
				}
				reqLogger.Info("service deleted")
			}(obj)
		},
	})
	informer.Run(stopper)
}


--- 

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
	updatedService, err := k8sClient.CoreV1().Services(service.Namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	updatedService.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: ip}}
	_, err = k8sClient.CoreV1().Services(updatedService.Namespace).UpdateStatus(context.TODO(), updatedService, metav1.UpdateOptions{})
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

// for type clusterIP or NodePort
func PatchFakeServiceToSetIP(k8sClient ExtendedClient, service *corev1.Service, ip string) (name string, error error) {
	service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: ip}}
	serviceName := service.Name
	_, err := k8sClient.CoreV1().Services(service.Namespace).UpdateStatus(context.TODO(), service, metav1.UpdateOptions{})
	return serviceName, err
}
