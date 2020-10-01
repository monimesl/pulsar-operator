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

package v1alpha1

import (
	"github.com/alphashaw/operator-pkg/webhooks"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var pulsarproxylog = logf.Log.WithName("pulsarproxy-resource")

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-pulsar-skulup-com-v1alpha1-pulsarproxy,mutating=true,failurePolicy=fail,groups=pulsar.skulup.com,resources=pulsarproxies,verbs=create;update,versions=v1alpha1,name=mpulsarproxy.kb.io

var _ webhook.Defaulter = &PulsarProxy{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *PulsarProxy) Default() {
	pulsarproxylog.Info("default", "name", in.Name)
	in.setSpecDefaults()
	in.setStatusDefaults()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-pulsar-skulup-com-v1alpha1-pulsarproxy,mutating=false,failurePolicy=fail,groups=pulsar.skulup.com,resources=pulsarproxies,versions=v1alpha1,name=vpulsarproxy.kb.io

var _ webhook.Validator = &PulsarProxy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarProxy) ValidateCreate() error {
	pulsarproxylog.Info("validate create", "name", in.Name)
	return in.validateProxy()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarProxy) ValidateUpdate(old runtime.Object) error {
	pulsarproxylog.Info("validate update", "name", in.Name)
	return in.validateProxy()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarProxy) ValidateDelete() error {
	pulsarproxylog.Info("validate delete", "name", in.Name)
	return nil
}

func (in *PulsarProxy) validateProxy() error {
	return webhooks.Validate(in.GroupVersionKind(), in.GetName(),
		func(list *webhooks.ErrorList) {
		})
}
