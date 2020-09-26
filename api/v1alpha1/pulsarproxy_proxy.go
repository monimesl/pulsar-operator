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
)

// +k8s:openapi-gen=true

// Proxy represents a pulsar proxy in the cluster
type Proxy struct {
	// Configs defines the configurations in `conf/proxy.conf` files in pulsar
	Configs string `json:"configs,omitempty"`

	// Image defines the container image to use. It defaults to apachepulsar/pulsar:latest
	Image pkg.Image `json:"image,omitempty"`

	// Labels defines the labels to attach to the broker deployment
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations defines the annotations to attach to the broker deployment
	Annotations map[string]string `json:"annotations,omitempty"`

	// PodConfig defines common configuration for the broker pods
	PodConfig pkg.PodConfig `json:"pod,omitempty"`
}

// Generate the labels of the proxy pod
func (in Proxy) GeneratePodLabels() map[string]string {
	return in.PodConfig.Labels
}

func (in *Proxy) setDefaults() (changed bool) {
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
