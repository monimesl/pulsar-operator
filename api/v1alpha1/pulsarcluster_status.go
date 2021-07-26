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

package v1alpha1

// ClusterStage represents the stage of the pulsar broker cluster
type ClusterStage string

const (
	// ClusterStageInitialized - cluster object is created but statefulset not created
	ClusterStageInitialized = "Initialized"
	// ClusterStageLaunching - cluster is initialized and the pods have been created but not ready
	ClusterStageLaunching = "Launched"
	// ClusterStageRunning - cluster is launched and running
	ClusterStageRunning = "Running"
)

// PulsarClusterStatus defines the observed state of PulsarCluster
type PulsarClusterStatus struct {

	// Replicas is the number of desired bookkeeper nodes in the cluster
	// +optional
	Replicas int32 `json:"replicas"`

	// CurrentReplicas is the number of current bookkeeper nodes in the cluster
	// +optional
	CurrentReplicas int32 `json:"currentReplicas"`

	// ReadyReplicas is the number of ready bookkeeper nodes in the cluster
	// +optional
	ReadyReplicas int32 `json:"readyReplicas"`

	// Metadata defines the metadata status of the cluster
	// +optional
	Metadata Metadata `json:"metadata,omitempty"`
}

// Metadata defines the metadata status of the cluster
type Metadata struct {
	Stage                 ClusterStage `json:"stage,omitempty"`
	ServiceMonitorVersion *string      `json:"serviceMonitorVersion,omitempty"`
}

func (in *PulsarClusterStatus) setDefaults() (changed bool) {
	if in.Metadata.Stage == "" {
		in.Metadata.Stage = ClusterStageInitialized
	}
	return
}
