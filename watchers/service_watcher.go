package watchers

import (
	"errors"
	"github.com/beopencloud/network-watcher/utils"

	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"net/http"

	// We need this import to load the GCP auth plugin which is required to authenticate against GKE clusters.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var serviceWatcherLogger = logf.Log.WithName("service_watcher")

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
				res, err := utils.PostRequestToAPI(utils.SERVICE_CREATE_EVENT_URL, service)
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
				res, err := utils.PutRequestToAPI(utils.SERVICE_UPDATE_EVENT_URL, service)
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
