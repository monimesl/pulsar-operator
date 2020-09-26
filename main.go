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
	"github.com/skulup/operator-pkg/configs"
	"github.com/skulup/operator-pkg/reconcilers"
	pulsarv1alpha1 "github.com/skulup/pulsar-operator/api/v1alpha1"
	"github.com/skulup/pulsar-operator/controllers/pulsarcluster"
	"github.com/skulup/pulsar-operator/controllers/pulsarmanager"
	"github.com/skulup/pulsar-operator/controllers/pulsarproxy"
	"github.com/skulup/pulsar-operator/pkg"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(pulsarv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	config, options := configs.GetManagerParams(scheme, "pulsar-operator", pkg.Domain)
	mgr, err := manager.New(config, options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	if err = reconcilers.Configure(mgr,
		&pulsarproxy.Reconciler{},
		&pulsarcluster.Reconciler{},
		&pulsarmanager.Reconciler{}); err != nil {
		setupLog.Error(err, "reconciler config error")
		os.Exit(1)
	}
	log.Fatal(mgr.Start(ctrl.SetupSignalHandler()))
}
