package watchers

import (
	"errors"
	"github.com/beopencloud/network-watcher/utils"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"net/http"
	//"strings"
	//"log"
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
					fmt.Println("Error 11", err)
					return
				}

				if service.Spec.Type == "ClusterIP" && service.Spec.Selector["run"] == "my-nginx" {
					_, err := utils.PatchFakeServiceToSetIP(k8sClient, service, ip)
					if err != nil {
						fmt.Println("Error ===========22", err)
						return
					}
				} else if service.Spec.Type == "LoadBalancer" && service.Spec.Selector["run"] == "loadbalancer" {
					listService, err := k8sClient.CoreV1().Services(service.Namespace).List(context.TODO(), metav1.ListOptions{})
					if err != nil {
						fmt.Println("Error 33==================", err)
						return
					}
					var name string
					var namespace string
					for _, v := range listService.Items {
						if v.Spec.Type == "ClusterIP" && v.Spec.Selector["run"] == "my-nginx" {
							name = v.Name
							namespace = v.Namespace
							fmt.Println(name, "", namespace)
						}
					}

					fakeService, err := k8sClient.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
					if err != nil {
						fmt.Println("Error 44======================", err)
					}
					err = utils.DeleteFakeService(k8sClient, fakeService)
					if err != nil {
						fmt.Println("Error 55=====================", err)
						return
					}
					err = utils.SetLoabBalancerIP(k8sClient, service, ip)
					if err != nil {
						fmt.Println("Error 66=====================", err)
						return
					}
				} else {
					fmt.Println("Not Found")
				}

				res, err := utils.PostRequestToAPI(service)
				if err != nil {
					fmt.Println("Error Post TO API", err)
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
					fmt.Println("Error", err)
					return
				}
				err = utils.SetLoabBalancerIP(k8sClient, service, ip)
				if err != nil {
					fmt.Println("Error", err)
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
					fmt.Println("Error", err)
					return
				}

				res, err := utils.DeleteRequestToAPI(utils.SERVICE_DELETE_EVENT_URL + "?kind=service&name=" + service.Name + "&namespace=" + service.Namespace)
				if err != nil {
					reqLogger.Error(err, "Error to send service delete event to API")
					return
				}

				if service.Spec.Selector["run"] == "loadbalancer" && service.Spec.Type == "LoadBalancer" {
					serviceAdd := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "fake-service",
							Namespace: service.Namespace,
							Labels: map[string]string{
								"run": "my-nginx",
							},
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{
								{Port: 80},
							},
							Selector:  map[string]string{"run": "my-nginx"},
							ClusterIP: service.Spec.ClusterIP,
						},
					}
					newFakeService, err := k8sClient.CoreV1().Services(service.Namespace).Create(context.TODO(), serviceAdd, metav1.CreateOptions{})
					if err != nil {
						fmt.Println("ErrorDeleting........", err)
						return
					}
					_, err = utils.PatchFakeServiceToSetIP(k8sClient, newFakeService, ip)
					if err != nil {
						fmt.Println("Error====================", err)
						return
					}

				}
				/* else if service.Spec.Selector["run"] == "my-nginx" && service.Name == "fake-service" {
					newService := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "fake-service",
							Namespace: service.Namespace,
							Labels: map[string]string{
								"run": "my-nginx",
							},
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{
								{Port: 80},
							},
							Selector:  map[string]string{"run": "my-nginx"},
							ClusterIP: "10.0.74.43",
						},
					}
					newFakeService, err := k8sClient.CoreV1().Services(service.Namespace).Create(context.TODO(), newService, metav1.CreateOptions{})
					if err != nil {
						fmt.Println("ErrorDeleting........", err)
						return
					}
					_, err = utils.PatchFakeServiceToSetIP(k8sClient, newFakeService, ip)
					if err != nil {
						fmt.Println("Error====================", err)
						return
					}
				} else {
					fmt.Println("No thinks to do !")
				}
				*/
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
