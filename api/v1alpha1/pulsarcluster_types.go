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

import (
	"fmt"
	"github.com/monimesl/operator-helper/basetype"
	"github.com/monimesl/operator-helper/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
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

func (in *PulsarCluster) nameHasPLSIndicator() bool {
	return strings.Contains(in.Name, "pls") || strings.Contains(in.Name, "pulsar")
}

func (in *PulsarCluster) generateName() string {
	if in.nameHasPLSIndicator() {
		return in.Name
	}
	return fmt.Sprintf("%s-pls", in.GetName())
}

// SetSpecDefaults set the defaults for the cluster spec and returns true otherwise false
func (in *PulsarCluster) SetSpecDefaults() bool {
	return in.Spec.setDefaults()
}

// SetStatusDefaults set the defaults for the cluster status and returns true otherwise false
func (in *PulsarCluster) SetStatusDefaults() bool {
	return in.Status.setDefaults()
}

// ConfigMapName defines the name of the configmap object
func (in *PulsarCluster) ConfigMapName() string {
	return in.generateName()
}

// StatefulSetName defines the name of the statefulset object
func (in *PulsarCluster) StatefulSetName() string {
	return in.generateName()
}

// ClientServiceName defines the name of the client service object
func (in *PulsarCluster) ClientServiceName() string {
	return in.generateName()
}

// HeadlessServiceName defines the name of the headless service object
func (in *PulsarCluster) HeadlessServiceName() string {
	return fmt.Sprintf("%s-headless", in.ClientServiceName())
}

// ClientServiceFQDN defines the FQDN of the client service object
func (in *PulsarCluster) ClientServiceFQDN() string {
	return fmt.Sprintf("%s.%s.svc.%s", in.ClientServiceName(), in.Namespace, in.Spec.ClusterDomain)
}

// ClientHeadlessServiceFQDN defines the FQDN of the client headless service object
func (in *PulsarCluster) ClientHeadlessServiceFQDN() string {
	return fmt.Sprintf("%s.%s.svc.%s", in.HeadlessServiceName(), in.Namespace, in.Spec.ClusterDomain)
}

func (in *PulsarCluster) CreateLabels(addPodLabels bool, more map[string]string) map[string]string {
	return in.Spec.createLabels(in.Name, addPodLabels, more)
}

// Image specifies the pulsar image to use
func (in *PulsarCluster) Image() basetype.Image {
	return basetype.Image{
		Repository: imageRepository,
		PullPolicy: in.Spec.ImagePullPolicy,
		Tag:        in.Spec.PulsarVersion,
	}
}

func (in *PulsarCluster) BrokersSetupPvcName() string {
	return fmt.Sprintf("broker-setup-%s", in.GetName())
}
