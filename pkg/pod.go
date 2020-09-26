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

package pkg

import (
	"github.com/skulup/operator-pkg/k8s"
	v1 "k8s.io/api/core/v1"
)

// +k8s:openapi-gen=true
// +kubebuilder:object:generate=true

// PodConfig defines the configurations of a kubernetes pod
type PodConfig struct {

	// Labels defines the labels to attach to the broker pod
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations defines the annotations to attach to the broker pod
	Annotations map[string]string `json:"annotations,omitempty"`

	// Compute Resources required by this container.
	// This field cannot be updated once the pod is created
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// List of environment variables to set in the container.
	// This field cannot be updated once the pod is created
	EnvVar []v1.EnvVar `json:"env,omitempty"`

	// List of sources to populate environment variables in the container.
	// The keys defined within a source must be a C_IDENTIFIER. All invalid keys
	// will be reported as an event when the container is starting. When a key exists in multiple
	// sources, the value associated with the last source will take precedence.
	// Values defined by an Env with a duplicate key will take precedence.
	// This field cannot be updated once the pod is created
	EnvFrom []v1.EnvFromSource `json:"envFrom,omitempty"`

	// Affinity defines the pod's scheduling constraints
	Affinity v1.Affinity `json:"affinity,omitempty"`

	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Optional duration in seconds the pod may be active on the node relative to
	// StartTime before the system will actively try to mark it failed and kill associated containers.
	// Value must be a positive integer.
	ActiveDeadlineSeconds int64 `json:"activeDeadlineSeconds,omitempty"`
	// Restart policy for all containers within the pod.
	// One of Always, OnFailure, Never.
	// Default to Always.
	RestartPolicy v1.RestartPolicy `json:"restartPolicy,omitempty"`

	SecurityContext v1.PodSecurityContext `json:"securityContext,omitempty"`
}

// Generate the pod environment sources
func (in *PodConfig) GenerateEnvFrom(sources ...v1.EnvFromSource) []v1.EnvFromSource {
	envFrom := make([]v1.EnvFromSource, 0)
	copy(envFrom, sources)
	if in.EnvFrom != nil {
		envFrom = append(envFrom, in.EnvFrom...)
	}
	return envFrom
}

// Generate the pod environment variables
func (in *PodConfig) GenerateEnvVar(sources ...v1.EnvVar) []v1.EnvVar {
	envVar := append(make([]v1.EnvVar, 0), sources...)
	if in.EnvVar != nil {
		envVar = append(envVar, in.EnvVar...)
	}
	envVar = append(envVar, v1.EnvVar{
		Name: k8s.EnvVarPodIP,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	})
	return envVar
}
