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
	"log"

	"github.com/skulup/operator-pkg/configs"
	"github.com/skulup/operator-pkg/reconcilers"
	"github.com/skulup/operator-pkg/webhooks"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	pulsarv1alpha1 "github.com/skulup/pulsar-operator/api/v1alpha1"
	"github.com/skulup/pulsar-operator/controllers/pulsarcluster"
	"github.com/skulup/pulsar-operator/controllers/pulsarmanager"
	"github.com/skulup/pulsar-operator/controllers/pulsarproxy"
	"github.com/skulup/pulsar-operator/internal"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(pulsarv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	config, options := configs.GetManagerParams(scheme, internal.OperatorName, internal.Domain)
	mgr, err := manager.New(config, options)
	if err != nil {
		log.Fatalf("unable to start manager: %s", err)
	}
	if err = webhooks.Configure(mgr,
		&pulsarv1alpha1.PulsarProxy{},
		&pulsarv1alpha1.PulsarCluster{},
		&pulsarv1alpha1.PulsarManager{}); err != nil {
		log.Fatalf("webhook config error: %s", err)
	}
	if err = reconcilers.Configure(mgr,
		&pulsarproxy.Reconciler{},
		&pulsarcluster.Reconciler{},
		&pulsarmanager.Reconciler{}); err != nil {
		log.Fatalf("unable to start manager: %s", err)
	}
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("operator start error: %s", err)
	}
}
