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
var pulsarmanagerlog = logf.Log.WithName("pulsarmanager-resource")

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-pulsar-skulup-com-v1alpha1-pulsarmanager,mutating=true,failurePolicy=fail,groups=pulsar.wirelimited.com,resources=pulsarmanagers,verbs=create;update,versions=v1alpha1,name=mpulsarmanager.kb.io

var _ webhook.Defaulter = &PulsarManager{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *PulsarManager) Default() {
	pulsarmanagerlog.Info("default", "name", in.Name)
	in.setSpecDefaults()
	in.setStatusDefaults()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-pulsar-skulup-com-v1alpha1-pulsarmanager,mutating=false,failurePolicy=fail,groups=pulsar.wirelimited.com,resources=pulsarmanagers,versions=v1alpha1,name=vpulsarmanager.kb.io

var _ webhook.Validator = &PulsarManager{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarManager) ValidateCreate() error {
	pulsarmanagerlog.Info("validate create", "name", in.Name)
	return in.validateManager()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarManager) ValidateUpdate(old runtime.Object) error {
	pulsarmanagerlog.Info("validate update", "name", in.Name)
	return in.validateManager()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarManager) ValidateDelete() error {
	pulsarmanagerlog.Info("validate delete", "name", in.Name)
	return nil
}

func (in *PulsarManager) validateManager() error {
	return webhooks.Validate(in.GroupVersionKind(), in.GetName(),
		func(list *webhooks.ErrorList) {
		})
}
