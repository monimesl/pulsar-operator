/*
 * Copyright 2021 - now, the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//nolint:dupl
package v1alpha1

import (
	"github.com/monimesl/operator-helper/config"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupWebhookWithManager needed for webhook test suite
func (in *PulsarProxy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-pulsar-monime-sl-v1alpha1-pulsarproxy,mutating=true,failurePolicy=fail,sideEffects=None,groups=pulsar.monime.sl,resources=pulsarproxies,verbs=create;update,versions=v1alpha1,name=mpulsarproxy.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &PulsarProxy{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *PulsarProxy) Default() {
	config.RequireRootLogger().Info("[Webhook] Setting defaults", "name", in.Name)
	in.SetSpecDefaults()
	in.SetStatusDefaults()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-pulsar-monime-sl-v1alpha1-pulsarproxy,mutating=false,failurePolicy=fail,sideEffects=None,groups=pulsar.monime.sl,resources=pulsarproxies,verbs=create;update,versions=v1alpha1,name=vpulsarproxy.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &PulsarProxy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarProxy) ValidateCreate() (admission.Warnings, error) {
	config.RequireRootLogger().Info("[validate create]", "name", in.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarProxy) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	config.RequireRootLogger().Info("[validate update]", "name", in.Name)
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *PulsarProxy) ValidateDelete() (admission.Warnings, error) {
	config.RequireRootLogger().Info("[validate delete]", "name", in.Name)
	return nil, nil
}
