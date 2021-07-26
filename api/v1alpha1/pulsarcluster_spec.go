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
	"github.com/monimesl/operator-helper/basetype"
	"github.com/monimesl/operator-helper/k8s"
	"github.com/monimesl/operator-helper/k8s/pod"
	"github.com/monimesl/operator-helper/operator/prometheus"
	"github.com/monimesl/pulsar-operator/internal"
	v1 "k8s.io/api/core/v1"
)

const (
	imageRepository = "monime/pulsar"
	defaultImageTag = "latest"
)

const (
	ClientPortName     = "client-port"
	ClientTLSPortName  = "client-tls-port"
	WebPortName        = "web-port"
	WebTLSPortName     = "web-tls-port"
	MetricsPortName    = "metrics-port"
	ServiceMetricsPath = "/metrics"
)

const (
	minimumClusterSize   = 3
	defaultClusterDomain = "cluster.local"
)

const (
	defaultClientPort    = 6650
	defaultClientTLSPort = 6651
	defaultWebPort       = 8080
	defaultWebTLSPort    = 8443
)

var (
	defaultTerminationGracePeriod int64 = 30
	defaultClusterSize                  = int32(minimumClusterSize)
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarClusterSpec defines the desired state of PulsarCluster
type PulsarClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ZookeeperServers specifies the hostname/IP address and port in the format "hostname:port".
	// +kubebuilder:validation:Required
	ZookeeperServers string `json:"zookeeperServers"`
	// ZookeeperServers specifies the hostname/IP address and port in the format "hostname:port".
	// +optional
	ConfigurationStoreServers string `json:"configurationStoreServers"`
	// PulsarVersion defines the version of bookkeeper to use
	// +optional
	PulsarVersion string `json:"pulsarVersion,omitempty"`
	// ImagePullPolicy describes a policy for if/when to pull the image
	// +optional
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// +kubebuilder:validation:Minimum=0
	Size *int32 `json:"size,omitempty"`
	// MaxUnavailableNodes defines the maximum number of nodes that
	// can be unavailable as per kubernetes PodDisruptionBudget
	// Default is 1.
	// +optional
	MaxUnavailableNodes int32  `json:"maxUnavailableNodes"`
	Ports               *Ports `json:"ports,omitempty"`
	// BrokerConfig defines the Bookkeeper configurations to override the bk_server.conf
	// https://github.com/apache/bookkeeper/tree/master/docker#configuration
	// +optional
	BrokerConfig map[string]string `json:"brokerConfig"`
	// PodConfig defines common configuration for the bookkeeper pods
	// +optional
	PodConfig basetype.PodConfig `json:"podConfig,omitempty"`
	// ProbeConfig defines the probing settings for the bookkeeper containers
	// +optional
	ProbeConfig *pod.Probes `json:"probeConfig,omitempty"`
	// MetricConfig
	// +optional
	MetricConfig *prometheus.MetricSpec `json:"metricConfig,omitempty"`
	// Env defines environment variables for the bookkeeper statefulset pods
	Env []v1.EnvVar `json:"env,omitempty"`

	// Labels defines the labels to attach to the bookkeeper deployment
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations defines the annotations to attach to the bookkeeper statefulset and services
	Annotations map[string]string `json:"annotations,omitempty"`

	// ClusterDomain defines the cluster domain for the cluster
	// It defaults to cluster.local
	ClusterDomain string `json:"clusterDomain,omitempty"`
}

type Ports struct {
	// +kubebuilder:validation:Minimum=1
	Client int32 `json:"client,omitempty"`
	// +kubebuilder:validation:Minimum=1
	ClientTLS int32 `json:"clientTLS,omitempty"`
	// +kubebuilder:validation:Minimum=1
	Web int32 `json:"web,omitempty"`
	// +kubebuilder:validation:Minimum=1
	WebTLS int32 `json:"WebTLS,omitempty"`
}

func (in *Ports) setDefaults() (changed bool) {
	if in.Client == 0 {
		changed = true
		in.Client = defaultClientPort
	}
	if in.ClientTLS == 0 {
		changed = true
		in.ClientTLS = defaultClientTLSPort
	}
	if in.Web == 0 {
		changed = true
		in.Web = defaultWebPort
	}
	if in.WebTLS == 0 {
		changed = true
		in.WebTLS = defaultWebTLSPort
	}
	return
}

// setDefaults set the defaults for the cluster spec and returns true otherwise false
func (in *PulsarClusterSpec) setDefaults() (changed bool) {
	if in.PulsarVersion == "" {
		changed = true
		in.PulsarVersion = defaultImageTag
	}
	if in.ImagePullPolicy == "" {
		changed = true
		in.ImagePullPolicy = v1.PullIfNotPresent
	}
	if in.Size == nil {
		changed = true
		size := &defaultClusterSize
		in.Size = size
	}
	if in.MaxUnavailableNodes < 0 {
		changed = true
		in.MaxUnavailableNodes = 1
	}
	if in.ClusterDomain == "" {
		changed = true
		in.ClusterDomain = defaultClusterDomain
	}
	if in.BrokerConfig == nil {
		in.BrokerConfig = map[string]string{}
	}
	if in.Ports == nil {
		in.Ports = &Ports{}
		in.Ports.setDefaults()
		changed = true
	} else if in.Ports.setDefaults() {
		changed = true
	}
	if in.ProbeConfig == nil {
		changed = true
		in.ProbeConfig = &pod.Probes{}
		in.ProbeConfig.SetDefault()
	} else if in.ProbeConfig.SetDefault() {
		changed = true
	}
	if in.PodConfig.TerminationGracePeriodSeconds == nil {
		changed = true
		in.PodConfig.TerminationGracePeriodSeconds = &defaultTerminationGracePeriod
	}
	return
}

func (in *PulsarClusterSpec) createLabels(clusterName string, addPodLabels bool, more map[string]string) map[string]string {
	ls := in.Labels
	if ls == nil {
		ls = map[string]string{}
	}
	if addPodLabels {
		for k, v := range in.PodConfig.Labels {
			ls[k] = v
		}
	}
	for k, v := range more {
		ls[k] = v
	}
	ls[k8s.LabelAppManagedBy] = internal.OperatorName
	ls[k8s.LabelAppName] = clusterName
	return ls
}
