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
	"github.com/skulup/pulsar-operator/pkg"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultRepository = "apachepulsar/pulsar"
const defaultTag = "latest"

// +k8s:openapi-gen=true
// Broker represents a pulsar broker in the cluster
type Broker struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ZookeeperServers          string `json:"zookeeperServers,omitempty"`
	ConfigurationStoreServers string `json:"configurationStoreServers,omitempty"`
	// Configs defines the configurations in `conf/broker.conf` files in pulsar
	Configs string `json:"configs,omitempty"`
	// Image defines the container image to use. It defaults to apachepulsar/pulsar:latest
	Image pkg.Image `json:"image,omitempty"`

	// Labels defines the labels to attach to the broker deployment
	Labels map[string]string `json:"labels,omitempty"`

	LabelSelector v1.LabelSelector `json:"selector,omitempty"`

	// Annotations defines the annotations to attach to the broker deployment
	Annotations map[string]string `json:"annotations,omitempty"`

	// PodConfig defines common configuration for the broker pods
	PodConfig pkg.PodConfig `json:"pod,omitempty"`
}

// Generate the labels of the broker pod and adds
// a `cluster` label with the value of the cluster name
func (in Broker) GeneratePodLabels(clusterName string) map[string]string {
	labels := map[string]string{}
	for k, v := range in.PodConfig.Labels {
		labels[k] = v
	}
	labels[pkg.LabelCluster] = clusterName
	return labels
}

// Set the defaults properties of the broker and returns
// true if any property was set otherwise false
func (in *Broker) SetDefaults() (changed bool) {
	if in.ConfigurationStoreServers == "" {
		changed = true
		in.ConfigurationStoreServers = in.ZookeeperServers
	}
	if in.Image.Repository == "" {
		changed = true
		in.Image.Repository = defaultRepository
	}
	if in.Image.Tag == "" {
		changed = true
		in.Image.Tag = defaultTag
	}
	if in.Image.PullPolicy == "" {
		changed = true
		in.Image.PullPolicy = v12.PullIfNotPresent
	}
	return
}
