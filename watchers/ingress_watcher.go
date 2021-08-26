package watchers

import (
	"errors"
	"github.com/beopencloud/network-watcher/utils"
	"io/ioutil"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"net/http"

	// We need this import to load the GCP auth plugin which is required to authenticate against GKE clusters.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"encoding/base64"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"context"
	"strings"
)

var ingressWatcherLogger = logf.Log.WithName("ingress_watcher")

// ++
// +
// Voici l'implémentation du watcher d'Ingress.
// vous pouvez voir dans le code les fonctions de callback appelées, dépendance des events add, update et delete Ingress.
// Au niveau de chaque fonction de callback, on récupère l'Ingress concerné et on envoie l'event à l'API.
// par exemple, lors de la création d'un service, nous obtenons l'objet de l'Ingress nouvellement créé qu'on
// envoie à l'API pour lui notifier la creation du nouveau service.
// +
// ++
func ingressWatch(k8sClient utils.ExtendedClient, stopper chan struct{}) {
	factory := informers.NewSharedInformerFactory(k8sClient, 0)
	informer := factory.Extensions().V1beta1().Ingresses().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			go func(obj interface{}) {
				ingress := obj.(*v1beta1.Ingress)
				reqLogger := ingressWatcherLogger.WithValues("service", ingress.Name, "namespace", ingress.Namespace)
				watch, err := utils.CheckNamespaceAutoGen(k8sClient, ingress.Namespace)
				if !watch || err != nil {
					return
				}
				secret, err := k8sClient.CoreV1().Secrets(ingress.Namespace).Get(context.TODO(),"secret-mock", metav1.GetOptions{})
				if err != nil{
					panic(err.Error())
				}
				username := string(secret.Data["username"])
				password := string(secret.Data["password"])
				urlBase      := string(secret.Data["url-ingress"])
				temp := strings.Split(urlBase, "\n")
				url :=strings.Join(temp, "")+"post"
				credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
				res, err := utils.PostRequestToAPI(url, credentials, ingress)
				if err != nil {
					reqLogger.Error(err, "Error to send ingress create event to API")
					return
				}
				if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
					responseBody, _ := ioutil.ReadAll(res.Body)
					reqLogger.Error(errors.New("Error to send ingress create event to API"), string(responseBody), "StatusCode", res.StatusCode)
					return
				}
				reqLogger.Info("ingress created")
			}(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			go func(oldObj, newObj interface{}) {
				ingress := newObj.(*v1beta1.Ingress)
				reqLogger := ingressWatcherLogger.WithValues("service", ingress.Name, "namespace", ingress.Namespace)
				watch, err := utils.CheckNamespaceAutoGen(k8sClient, ingress.Namespace)
				if !watch || err != nil {
					return
				}
				res, err := utils.PutRequestToAPI(utils.INGRESS_UPDATE_EVENT_URL, ingress)
				if err != nil {
					reqLogger.Error(err, "Error to send ingress update event to API")
					return
				}
				if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
					responseBody, _ := ioutil.ReadAll(res.Body)
					reqLogger.Error(errors.New("Error to send ingress update event to API"), string(responseBody), "StatusCode", res.StatusCode)
					return
				}
				reqLogger.Info("ingress updated")
			}(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			go func(obj interface{}) {
				ingress := obj.(*v1beta1.Ingress)
				reqLogger := ingressWatcherLogger.WithValues("service", ingress.Name, "namespace", ingress.Namespace)
				watch, err := utils.CheckNamespaceAutoGen(k8sClient, ingress.Namespace)
				if !watch || err != nil {
					return
				}
				res, err := utils.DeleteRequestToAPI(utils.INGRESS_DELETE_EVENT_URL + "?kind=ingress&name=" + ingress.Name + "&namespace=" + ingress.Namespace)
				if err != nil {
					reqLogger.Error(err, "Error to send ingress create event to API")
					return
				}
				if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
					responseBody, _ := ioutil.ReadAll(res.Body)
					reqLogger.Error(errors.New("Error to send ingress create event to API"), string(responseBody), "StatusCode", res.StatusCode)
					return
				}
				reqLogger.Info("ingress deleted")
			}(obj)
		},
	})
	informer.Run(stopper)
}
