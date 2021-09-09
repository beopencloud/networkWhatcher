package watchers

import (
	"errors"
	"github.com/beopencloud/network-watcher/utils"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
	"net/http"
	"time"

	//"strings"
	//"log"
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	"time"

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
	test := false
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
					fmt.Println("111 Error 11", err)
					return
				}

				if service.Spec.Type == "NodePort" && service.Labels["servicetype"] == "LoadBalancer" {
					// TODO Get fake-service and Delete fake-service
					listService, err := k8sClient.CoreV1().Services(service.Namespace).List(context.TODO(), metav1.ListOptions{})
					if err != nil {
						fmt.Println("111 Error 33==================", err)
						return
					}
					var fakeService  corev1.Service
					for _, v := range listService.Items {
						if v.Name == "fake-service" {
							fakeService = v
							fmt.Println("111 +++++++++", fakeService.Name, "", fakeService.Namespace)
						}
					}
					fmt.Println("111 A====A++++++", fakeService.Name)
					err = utils.DeleteFakeService(k8sClient, &fakeService)
					fmt.Println("111 Deleting.......")
					if err != nil {
						fmt.Println("111 Error 55=====================", err)
						return
					}
					test = false
					fmt.Println("111 Deleting.......done", test)
					//		time.Sleep(30 * time.Second)
					// TODO Patch service type to LoadBalancer et IP annotation
					err = utils.SetLoabBalancerIP(k8sClient, service, ip)
					if err != nil {
						fmt.Println("111 Error 66=====================", err)
						return
					}
				} else {
					fmt.Println("111 Not Found")
				}

				res, err := utils.PostRequestToAPI(service)
				if err != nil {
					fmt.Println("111 Error Post TO API", err)
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
				log.Println("DELETING SERVICE: ", service.Name)
				reqLogger := serviceWatcherLogger.WithValues("service", service.Name, "namespace", service.Namespace)
				watch, err := utils.CheckNamespaceAutoGen(k8sClient, service.Namespace)
				if !watch || err != nil {
					return
				}
				ip, err := utils.GetNamespaceIP(k8sClient, service.Namespace)
				if err != nil {
					fmt.Println("222 Error", err)
					return
				}

			getServices:
				services, err := k8sClient.CoreV1().Services(service.Namespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					fmt.Println("222 Error", err)
					return
				}
				for _, v := range services.Items {
					if v.Name == service.Name {
						log.Println("waiting for the complete deletion of service", service.Name, "...")
						time.Sleep(2*time.Second)
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
								{Port: 80},
							},
						},
					}
					newFakeService, err := k8sClient.CoreV1().Services(service.Namespace).Create(context.TODO(), serviceAdd, metav1.CreateOptions{})
					if err != nil {
						fmt.Println("222 Error creating fake-service........", err)
						return
					}
					_, err = utils.PatchFakeServiceToSetIP(k8sClient, newFakeService, ip)
					if err != nil {
						fmt.Println("222 Error====================", err)
						return
					}

				}

				res, err := utils.DeleteRequestToAPI(utils.SERVICE_DELETE_EVENT_URL + "?kind=service&name=" + service.Name + "&namespace=" + service.Namespace)
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
