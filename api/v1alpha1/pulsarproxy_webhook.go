/*
Copyright 2021.

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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var pulsarproxylog = logf.Log.WithName("pulsarproxy-resource")

func (r *PulsarProxy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-pulsar-monime-sl-v1alpha1-pulsarproxy,mutating=true,failurePolicy=fail,sideEffects=None,groups=pulsar.monime.sl,resources=pulsarproxies,verbs=create;update,versions=v1alpha1,name=mpulsarproxy.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &PulsarProxy{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *PulsarProxy) Default() {
	pulsarproxylog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-pulsar-monime-sl-v1alpha1-pulsarproxy,mutating=false,failurePolicy=fail,sideEffects=None,groups=pulsar.monime.sl,resources=pulsarproxies,verbs=create;update,versions=v1alpha1,name=vpulsarproxy.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &PulsarProxy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *PulsarProxy) ValidateCreate() error {
	pulsarproxylog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *PulsarProxy) ValidateUpdate(old runtime.Object) error {
	pulsarproxylog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *PulsarProxy) ValidateDelete() error {
	pulsarproxylog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
