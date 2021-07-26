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
	"github.com/monimesl/operator-helper/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	_ reconciler.Defaulting = &PulsarCluster{}
)

//+kubebuilder:object:root=true

// PulsarClusterList contains a list of PulsarCluster
type PulsarClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarCluster{}, &PulsarClusterList{})
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PulsarCluster is the Schema for the pulsarclusters API
type PulsarCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarClusterSpec   `json:"spec,omitempty"`
	Status PulsarClusterStatus `json:"status,omitempty"`
}

// SetSpecDefaults set the defaults for the cluster spec and returns true otherwise false
func (in *PulsarCluster) SetSpecDefaults() bool {
	return in.Spec.setDefaults()
}

// SetStatusDefaults set the defaults for the cluster status and returns true otherwise false
func (in *PulsarCluster) SetStatusDefaults() bool {
	return in.Status.setDefaults()
}
