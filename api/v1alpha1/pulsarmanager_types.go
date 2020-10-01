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
	"github.com/skulup/operator-helper/types"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	managerDefaultDbUserName = "pulsar"
	managerDefaultDbPassword = "pulsar"
	// ManagerDefaultSuperUsername defines the default username of the manager's superuser
	ManagerDefaultSuperUsername = "admin"
	// ManagerDefaultImageRepository defines the default pulsar-manager image repository
	ManagerDefaultImageRepository = "apachepulsar/pulsar-manager"
	// ManagerDefaultImageTag defines the default pulsar-manager image tag
	ManagerDefaultImageTag = "latest"
	// ManagerDefaultLogLevel defines the default pulsar-manager log level
	ManagerDefaultLogLevel = "DEBUG"
	// ManagerDefaultVolumeMountPath defines the default pulsar-manager postgreSQL data volume path
	ManagerDefaultVolumeMountPath = "/data"
)

// PulsarManagerSpec defines the desired state of PulsarManager
type PulsarManagerSpec struct {

	// Username defines the superuser username
	Username string `json:"username,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
	// Email defines the superuser email
	Email string `json:"email,omitempty"`

	// +kubebuilder:validation:MinLength=6
	// DbUsername defines the username of PostgreSQL
	DbUsername string `json:"dbUsername,omitempty"`

	// +kubebuilder:validation:MinLength=6
	// DbPassword defines the password of PostgreSQL
	DbPassword string `json:"dbPassword,omitempty"`

	// LogLevel defaults to DEBUG
	LogLevel string `json:"logLevel,omitempty"`

	// Image defines the container image to use. It defaults to apachepulsar/pulsar-manager:latest
	Image types.Image `json:"image,omitempty"`

	// Labels defines the labels to attach to the broker deployment
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations defines the annotations to attach to the broker deployment
	Annotations map[string]string `json:"annotations,omitempty"`

	// PodConfig defines common configuration for the broker pods
	PodConfig types.PodConfig `json:"pod,omitempty"`
}

// GeneratePodLabels generates the labels of the manager pod
func (in *PulsarManagerSpec) GeneratePodLabels() map[string]string {
	return in.Labels
}

// PulsarManagerStatus defines the observed state of PulsarManager
type PulsarManagerStatus struct {
	SuperUserAccountCreated bool `json:"superUserAccountCreated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PulsarManager is the Schema for the pulsarmanagers API
type PulsarManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarManagerSpec   `json:"spec,omitempty"`
	Status PulsarManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PulsarManagerList contains a list of PulsarManager
type PulsarManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarManager{}, &PulsarManagerList{})
}

func (in *PulsarManager) setSpecDefaults() {
	if in.Spec.Username == "" {
		in.Spec.Username = ManagerDefaultSuperUsername
	}
	if in.Spec.DbUsername == "" {
		in.Spec.DbUsername = managerDefaultDbUserName
	}
	if in.Spec.DbPassword == "" {
		in.Spec.DbPassword = managerDefaultDbPassword
	}
	if in.Spec.LogLevel == "" {
		in.Spec.LogLevel = ManagerDefaultLogLevel
	}
	if in.Spec.Image.Repository == "" {
		in.Spec.Image.Repository = ManagerDefaultImageRepository
	}
	if in.Spec.Image.Tag == "" {
		in.Spec.Image.Tag = ManagerDefaultImageTag
	}
	if in.Spec.Image.PullPolicy == "" {
		in.Spec.Image.PullPolicy = v12.PullIfNotPresent
	}
	return
}

func (in *PulsarManager) setStatusDefaults() bool {
	return false
}
