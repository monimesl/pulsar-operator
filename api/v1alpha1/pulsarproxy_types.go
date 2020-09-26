/*
 * Copyright 2020 Skulup Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarProxySpec defines the desired state of PulsarProxy
type PulsarProxySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// Size defines the number of broker in the cluster
	Count int32 `json:"count,omitempty"`

	// Proxy defines the desired state of proxy in the cluster
	Proxy Proxy `json:"proxy,omitempty"`
}

// PulsarProxyStatus defines the observed state of PulsarProxy
type PulsarProxyStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PulsarProxy is the Schema for the pulsarproxies API
type PulsarProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarProxySpec   `json:"spec,omitempty"`
	Status PulsarProxyStatus `json:"status,omitempty"`
}

// SetSpecDefaults set the defaults properties of the proxy spec and returns
// true if any property was set otherwise false
func (in *PulsarProxy) SetSpecDefaults() bool {
	return in.Spec.Proxy.setDefaults()
}

// SetStatusDefaults set the defaults properties of the proxy status and returns
// true if any property was set otherwise false
func (in *PulsarProxy) SetStatusDefaults() bool {
	return false
}

// +kubebuilder:object:root=true

// PulsarProxyList contains a list of PulsarProxy
type PulsarProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarProxy{}, &PulsarProxyList{})
}
