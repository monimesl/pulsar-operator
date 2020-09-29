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

// ClusterStage represents the stage of the pulsar broker cluster
type ClusterStage string

const (
	// The cluster has been created but not initialized
	ClusterStageInitializing ClusterStage = "ClusterInitializing"
	// The cluster is initialized and is being launched
	ClusterStageLaunching = "ClusterLaunching"
	// The cluster has ready and running
	ClusterStageRunning = "ClusterRunning"
)

// PulsarClusterSpec defines the desired state of PulsarCluster
type PulsarClusterSpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// Size defines the number of broker in the cluster
	Size int32 `json:"size,omitempty"`

	// Broker defines the desired state of brokers in the cluster
	Broker Broker `json:"broker,omitempty"`
}

// PulsarClusterStatus defines the observed state of PulsarCluster
type PulsarClusterStatus struct {
	Stage ClusterStage `json:"stage,omitempty"`
}

// Check if the cluster has completed initialization
func (in *PulsarCluster) IsInitialized() bool {
	return in.Status.Stage != ClusterStageInitializing
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PulsarCluster is the Schema for the pulsarclusters API
type PulsarCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarClusterSpec   `json:"spec,omitempty"`
	Status PulsarClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PulsarClusterList contains a list of PulsarCluster
type PulsarClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarCluster{}, &PulsarClusterList{})
}

func (in *PulsarCluster) setSpecDefaults() {
	if in.Spec.Size == 0 {
		in.Spec.Size = 1
	}
	in.Spec.Broker.setDefaults()
}

func (in *PulsarCluster) setStatusDefaults() {
	if in.Status.Stage == "" {
		in.Status.Stage = ClusterStageInitializing
	}
}
