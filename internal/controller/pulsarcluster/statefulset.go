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

package pulsarcluster

import (
	"context"
	"fmt"
	"github.com/monimesl/operator-helper/k8s"
	"github.com/monimesl/operator-helper/k8s/pod"
	"github.com/monimesl/operator-helper/k8s/pvc"
	"github.com/monimesl/operator-helper/k8s/statefulset"
	"github.com/monimesl/operator-helper/reconciler"
	"github.com/monimesl/pulsar-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	dataVolumeMouthPath = "/data"
)

// ReconcileStatefulSet reconcile the statefulset of the specified cluster
func ReconcileStatefulSet(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	if cluster.Status.Metadata.Stage != v1alpha1.ClusterStageInitialized {
		return nil
	}
	sts := &v1.StatefulSet{}
	return ctx.GetResource(types.NamespacedName{
		Name:      cluster.StatefulSetName(),
		Namespace: cluster.Namespace,
	}, sts,
		// Found
		func() error {
			if shouldUpdateStatefulSet(cluster.Spec, sts) {
				if err := updateStatefulset(ctx, sts, cluster); err != nil {
					return err
				}
			}
			return nil
		},
		// Not Found
		func() error {
			sts = createStatefulSet(cluster)
			if err := ctx.SetOwnershipReference(cluster, sts); err != nil {
				return err
			}
			ctx.Logger().Info("Creating the pulsar broker statefulset.",
				"StatefulSet.Name", sts.GetName(),
				"StatefulSet.Namespace", sts.GetNamespace())
			if err := ctx.Client().Create(context.TODO(), sts); err != nil {
				return err
			}
			ctx.Logger().Info("StatefulSet creation success.",
				"StatefulSet.Name", sts.GetName(),
				"StatefulSet.Namespace", sts.GetNamespace())
			return nil
		})
}

func shouldUpdateStatefulSet(spec v1alpha1.PulsarClusterSpec, sts *v1.StatefulSet) bool {
	if *spec.Size != *sts.Spec.Replicas {
		return true
	}
	if spec.PulsarVersion != sts.Labels[k8s.LabelAppVersion] {
		return true
	}
	return false
}

func updateStatefulset(ctx reconciler.Context, sts *v1.StatefulSet, cluster *v1alpha1.PulsarCluster) error {
	sts.Spec.Replicas = cluster.Spec.Size
	sts.Labels = cluster.GenerateLabels(true)
	sts.Spec.Selector.MatchLabels = getBrokerSelectorLabels(cluster, true)
	ctx.Logger().Info("Updating the pulsar broker  statefulset.",
		"StatefulSet.Name", sts.GetName(),
		"StatefulSet.Namespace", sts.GetNamespace(), "NewReplicas", cluster.Spec.Size)
	return ctx.Client().Update(context.TODO(), sts)
}

func createStatefulSet(c *v1alpha1.PulsarCluster) *v1.StatefulSet {
	pvcs := createPersistentVolumeClaims(c)
	brokerSelectorLabels := getBrokerSelectorLabels(c, true)
	templateSpec := createPodTemplateSpec(c, brokerSelectorLabels)
	spec := statefulset.NewSpec(*c.Spec.Size, c.HeadlessServiceName(), brokerSelectorLabels, pvcs, templateSpec)
	sts := statefulset.New(c.Namespace, c.StatefulSetName(), c.GenerateLabels(true), spec)
	sts.Annotations = c.GenerateAnnotations()
	return sts
}

func createPodTemplateSpec(c *v1alpha1.PulsarCluster, labels map[string]string) v12.PodTemplateSpec {
	return v12.PodTemplateSpec{
		ObjectMeta: pod.NewMetadata(c.Spec.PodConfig, "",
			c.StatefulSetName(), labels,
			c.GenerateAnnotations()),
		Spec: createPodSpec(c),
	}
}

func createPodSpec(c *v1alpha1.PulsarCluster) v12.PodSpec {
	setupEnv := []v12.EnvVar{
		{Name: "PULSAR_VERSION", Value: c.Spec.PulsarVersion},
		{Name: "PULSAR_CONNECTORS", Value: generateConnectorString(c)},
		{Name: "PULSAR_DATA_DIRECTORY", Value: dataVolumeMouthPath},
	}
	volumeMounts := []v12.VolumeMount{
		{Name: c.BrokersDataPvcName(), MountPath: dataVolumeMouthPath},
	}
	initContainers := []v12.Container{
		{
			Name: "broker-setup",
			Image: fmt.Sprintf("%s:%s",
				v1alpha1.BrokerSetupImageRepository,
				v1alpha1.DefaultBrokerSetupImageVersion),
			ImagePullPolicy: v1alpha1.DefaultBrokerSetupImagePullPolicy,
			VolumeMounts:    volumeMounts,
			Env:             setupEnv,
		},
	}
	envs := processEnvVars(c.Spec.PodConfig.Spec.Env)
	envs = append(envs, v12.EnvVar{
		Name: "PULSAR_DATA_DIRECTORY", Value: dataVolumeMouthPath,
	})
	probePort := c.Spec.Ports.Web
	if probePort <= 0 {
		probePort = c.Spec.Ports.WebTLS
	}
	containers := []v12.Container{
		{
			Name:            "pulsar-broker",
			VolumeMounts:    volumeMounts,
			Ports:           createContainerPorts(c),
			Image:           c.Image().ToString(),
			ImagePullPolicy: c.Image().PullPolicy,
			Resources:       c.Spec.PodConfig.Spec.Resources,
			StartupProbe:    createStartupProbe(c.Spec, probePort),
			LivenessProbe:   createLivenessProbe(c.Spec, probePort),
			ReadinessProbe:  createReadinessProbe(c.Spec, probePort),
			Env:             pod.DecorateContainerEnvVars(true, envs...),
			EnvFrom: []v12.EnvFromSource{
				{
					ConfigMapRef: &v12.ConfigMapEnvSource{
						LocalObjectReference: v12.LocalObjectReference{
							Name: c.ConfigMapName(),
						},
					},
				},
			},
			Command: []string{"sh", "-c"},
			Args: []string{
				strings.Join([]string{
					"echo \"yeah\" > status",
					"rm -rf /pulsar/connectors",
					"cp -r \"$PULSAR_DATA_DIRECTORY/connectors\" /pulsar",
					"bin/apply-config-from-env.py conf/broker.conf",
					"bin/pulsar broker",
				}, "; "),
			},
		},
	}
	return pod.NewSpec(c.Spec.PodConfig, nil, initContainers, containers)
}

func generateConnectorString(c *v1alpha1.PulsarCluster) string {
	formats := make([]string, 0)
	formats = append(formats, c.Spec.Connectors.Builtin...)
	for i, connector := range c.Spec.Connectors.Custom {
		headers := ""
		for k, v := range connector.Headers {
			if headers == "" {
				headers += ";"
			}
			headers += fmt.Sprintf("%s:%s", k, v)
		}
		formats[i] = fmt.Sprintf("%s;%s", connector.URL, headers)
	}
	return strings.Join(formats, " ")
}

//nolint:dupl
func createContainerPorts(c *v1alpha1.PulsarCluster) []v12.ContainerPort {
	ports := c.Spec.Ports
	containerPorts := []v12.ContainerPort{{Name: v1alpha1.ClientPortName, ContainerPort: ports.Client}}
	if ports.ClientTLS > 0 {
		containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.ClientTLSPortName, ContainerPort: ports.ClientTLS})
	}
	if ports.Web > 0 {
		containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.WebPortName, ContainerPort: ports.Web})
	}
	if ports.WebTLS > 0 {
		containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.WebTLSPortName, ContainerPort: ports.WebTLS})
	}
	kop := c.Spec.KOP
	if kop.Enabled {
		if kop.PlainTextPort > 0 {
			containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.KopPlainTextPortName, ContainerPort: kop.PlainTextPort})
		}
		if kop.SecuredPort > 0 {
			containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.KopSecuredPortName, ContainerPort: kop.SecuredPort})
		}
	}
	return containerPorts
}

func createStartupProbe(spec v1alpha1.PulsarClusterSpec, port int32) *v12.Probe {
	return spec.ProbeConfig.Startup.ToK8sProbe(v12.ProbeHandler{
		HTTPGet: &v12.HTTPGetAction{
			Port: intstr.FromInt32(port),
			Path: "/status.html",
		},
	})
}

func createReadinessProbe(spec v1alpha1.PulsarClusterSpec, port int32) *v12.Probe {
	return spec.ProbeConfig.Readiness.ToK8sProbe(v12.ProbeHandler{
		HTTPGet: &v12.HTTPGetAction{
			Port: intstr.FromInt32(port),
			Path: "/status.html",
		},
	})
}

func createLivenessProbe(spec v1alpha1.PulsarClusterSpec, port int32) *v12.Probe {
	return spec.ProbeConfig.Readiness.ToK8sProbe(v12.ProbeHandler{
		HTTPGet: &v12.HTTPGetAction{
			Port: intstr.FromInt32(port),
			Path: "/status.html",
		},
	})
}

func createPersistentVolumeClaims(c *v1alpha1.PulsarCluster) []v12.PersistentVolumeClaim {
	return []v12.PersistentVolumeClaim{
		pvc.New(c.Namespace, c.BrokersDataPvcName(),
			c.GenerateLabels(true),
			*c.Spec.Persistence),
	}
}
