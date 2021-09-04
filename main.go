/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"github.com/beopencloud/network-watcher/watchers"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//	 +kubebuilder:scaffold:imports
	//"github.com/beopencloud/network-watcher/utils"
	//"k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/clientcmd"
	"log"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme

	// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete
	// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;delete
	// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;delete
}

// +
// Le code de base de l'operator a ete generer par l'operator sdk.
// Y'a juste le dockerfile et la fonction main qui ont ete un peu modifier et
// les packages watchers et utils qui on été entièrement implementer.
// +
func main() {
	//	kubeconf,_ := restConfig()

	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8089", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "618544d0.beopenit.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// ++
	// +
	// C'est ici qu'on fait appel aux watchers.
	// on utilise le goroutine pour que les watchers s'execute en background.
	// cette instruction est la seule ajouter au niveau de la fonction main. le reste est generer par l'operator sdk
	// +
	// ++
	watchers.Watch(mgr.GetConfig())

	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

}


