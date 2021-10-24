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
	"github.com/monimesl/operator-helper/k8s/annotation"
	"github.com/monimesl/operator-helper/k8s/pod"
	"github.com/monimesl/operator-helper/k8s/pvc"
	"github.com/monimesl/operator-helper/k8s/statefulset"
	"github.com/monimesl/operator-helper/reconciler"
	"github.com/monimesl/pulsar-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	brokerSetupPvcSize   = "3Gi"
	setupVolumeMouthPath = "/broker-setup"
)

// ReconcileStatefulSet reconcile the statefulset of the specified cluster
func ReconcileStatefulSet(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	sts := &v1.StatefulSet{}
	return ctx.GetResource(types.NamespacedName{
		Name:      cluster.StatefulSetName(),
		Namespace: cluster.Namespace,
	}, sts,
		// Found
		func() error {
			if *cluster.Spec.Size != *sts.Spec.Replicas {
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

func updateStatefulset(ctx reconciler.Context, sts *v1.StatefulSet, cluster *v1alpha1.PulsarCluster) error {
	sts.Spec.Replicas = cluster.Spec.Size
	ctx.Logger().Info("Updating the pulsar broker  statefulset.",
		"StatefulSet.Name", sts.GetName(),
		"StatefulSet.Namespace", sts.GetNamespace(), "NewReplicas", cluster.Spec.Size)
	return ctx.Client().Update(context.TODO(), sts)
}

func createStatefulSet(c *v1alpha1.PulsarCluster) *v1.StatefulSet {
	pvcs := createPersistentVolumeClaims(c)
	labels := c.CreateLabels(true, nil)
	templateSpec := createPodTemplateSpec(c, labels)
	spec := statefulset.NewSpec(*c.Spec.Size, c.HeadlessServiceName(), labels, pvcs, templateSpec)
	sts := statefulset.New(c.Namespace, c.StatefulSetName(), labels, spec)
	annotations := c.Spec.Annotations
	if c.Spec.MonitoringConfig.Enabled &&
		(c.Spec.Ports.Web > 0 || c.Spec.Ports.WebTLS > 0) {
		metricPort := c.Spec.Ports.Web
		if metricPort <= 0 {
			metricPort = c.Spec.Ports.WebTLS
		}
		annotations = annotation.DecorateForPrometheus(
			annotations, true, int(metricPort))
	}
	sts.Annotations = annotations
	return sts
}

func createPodTemplateSpec(c *v1alpha1.PulsarCluster, labels map[string]string) v12.PodTemplateSpec {
	return pod.NewTemplateSpec("", c.StatefulSetName(), labels, nil, createPodSpec(c))
}

func createPodSpec(c *v1alpha1.PulsarCluster) v12.PodSpec {
	setupEnv := []v12.EnvVar{
		{Name: "PULSAR_VERSION", Value: c.Spec.PulsarVersion},
		{Name: "PULSAR_CONNECTORS", Value: generateConnectorString(c)},
		{Name: "PULSAR_SETUP_DIRECTORY", Value: setupVolumeMouthPath},
	}
	volumeMounts := []v12.VolumeMount{
		{Name: c.BrokersSetupPvcName(), MountPath: setupVolumeMouthPath},
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
	envs := processEnvVars(c.Spec.Env)
	envs = append(envs, v12.EnvVar{
		Name: "PULSAR_SETUP_DIRECTORY", Value: setupVolumeMouthPath,
	})
	containers := []v12.Container{
		{
			Name:            "pulsar-broker",
			Ports:           createContainerPorts(c),
			Image:           c.Image().ToString(),
			ImagePullPolicy: c.Image().PullPolicy,
			StartupProbe:    createStartupProbe(c.Spec),
			LivenessProbe:   createLivenessProbe(c.Spec),
			ReadinessProbe:  createReadinessProbe(c.Spec),
			VolumeMounts:    volumeMounts,
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
					"cp -r \"$PULSAR_SETUP_DIRECTORY/connectors\" /pulsar",
					"bin/apply-config-from-env.py conf/broker.conf",
					"bin/pulsar broker",
				}, "; "),
			},
		},
	}
	spec := pod.NewSpec(c.Spec.PodConfig, nil, initContainers, containers)
	spec.TerminationGracePeriodSeconds = c.Spec.PodConfig.TerminationGracePeriodSeconds
	return spec
}

func generateConnectorString(c *v1alpha1.PulsarCluster) string {
	formats := make([]string, 0)
	for _, connector := range c.Spec.Connectors.Builtin {
		formats = append(formats, connector)
	}
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

func createStartupProbe(spec v1alpha1.PulsarClusterSpec) *v12.Probe {
	return spec.ProbeConfig.Startup.ToK8sProbe(v12.Handler{
		HTTPGet: &v12.HTTPGetAction{
			Port: intstr.FromInt(int(spec.Ports.Web)),
			Path: "/status.html",
		},
	})
}

func createReadinessProbe(spec v1alpha1.PulsarClusterSpec) *v12.Probe {
	return spec.ProbeConfig.Readiness.ToK8sProbe(v12.Handler{
		HTTPGet: &v12.HTTPGetAction{
			Port: intstr.FromInt(int(spec.Ports.Web)),
			Path: "/status.html",
		},
	})
}

func createLivenessProbe(spec v1alpha1.PulsarClusterSpec) *v12.Probe {
	return spec.ProbeConfig.Liveness.ToK8sProbe(v12.Handler{
		HTTPGet: &v12.HTTPGetAction{
			Port: intstr.FromInt(int(spec.Ports.Web)),
			Path: "/status.html",
		},
	})
}

func createPersistentVolumeClaims(c *v1alpha1.PulsarCluster) []v12.PersistentVolumeClaim {
	return []v12.PersistentVolumeClaim{
		pvc.New(c.Namespace, c.BrokersSetupPvcName(),
			c.CreateLabels(false, nil),
			v12.PersistentVolumeClaimSpec{
				Resources: v12.ResourceRequirements{
					Requests: map[v12.ResourceName]resource.Quantity{
						v12.ResourceStorage: resource.MustParse(brokerSetupPvcSize),
					}},
				AccessModes: []v12.PersistentVolumeAccessMode{v12.ReadWriteMany},
			}),
	}
}
