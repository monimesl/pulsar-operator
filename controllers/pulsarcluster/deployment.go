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

package pulsarcluster

import (
	"context"
	"fmt"
	"github.com/alphashaw/operator-pkg/k8s"
	"github.com/alphashaw/operator-pkg/k8s/deployment"
	"github.com/alphashaw/operator-pkg/reconciler"
	"github.com/alphashaw/pulsar-operator/api/v1alpha1"
	"github.com/alphashaw/pulsar-operator/internal"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	pulsarConfigEnvPrefix   = "PULSAR_PREFIX_"
	configEnvPyScriptVolume = "apply-env-config-script"
)

func reconcileDeployment(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	if cluster.IsInitialized() {
		dep := &v1.Deployment{}
		return ctx.GetResource(types.NamespacedName{
			Namespace: deploymentNamespace(cluster),
			Name:      deploymentName(cluster),
		}, dep,
			func() (err error) { // Deployment Found
				updated := false
				oldReplicaCount := *dep.Spec.Replicas
				if oldReplicaCount != cluster.Spec.Size {
					updated = true
					dep.Spec.Replicas = &cluster.Spec.Size
				}
				if updated {
					if err = ctx.Client().Update(context.TODO(), dep); err == nil {
						ctx.Logger().Info("Deployment replica set scale successfully. ",
							"Deployment.Name", dep.GetName(),
							"Deployment.Namespace", dep.GetNamespace(),
							"Old Deployment.ReplicaSize", oldReplicaCount,
							"New Deployment.ReplicaSize", cluster.Spec.Size)
					}
				}
				return err
			},
			func() (err error) { // Deployment not Found
				dep = createDeployment(cluster)
				if err = ctx.SetOwnershipReference(cluster, dep); err == nil {
					if err = ctx.Client().Create(context.TODO(), dep); err == nil {
						ctx.Logger().Info("Deployment creation success. ",
							"Deployment.Name", dep.GetName(),
							"Deployment.Namespace", dep.GetNamespace())
					}
				}
				return
			})
	}
	return nil
}

func isDeploymentReady(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) bool {
	return deployment.IsReady(ctx.Client(),
		deploymentNamespace(cluster),
		deploymentName(cluster),
		cluster.Spec.Size)
}

func deploymentNamespace(c *v1alpha1.PulsarCluster) string {
	return c.GetNamespace()
}

func deploymentName(c *v1alpha1.PulsarCluster) string {
	return fmt.Sprintf("%s-pulsar-broker", c.GetName())
}

func createDeployment(c *v1alpha1.PulsarCluster) *v1.Deployment {
	labels := internal.GenerateLabels(internal.Broker, c.Spec.Broker.GeneratePodLabels(c.GetName()))
	return deployment.New(deploymentNamespace(c), deploymentName(c), c.Spec.Broker.Labels,
		v1.DeploymentSpec{
			Replicas: &c.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: createDeploymentPodTemplateSpec(c, labels),
		})
}

func createDeploymentPodTemplateSpec(c *v1alpha1.PulsarCluster, podTemplateLabels map[string]string) v12.PodTemplateSpec {
	ConfigMapDefaultMode := int32(0744)
	podCfg := c.Spec.Broker.PodConfig
	var activeDeadlineSeconds *int64
	if podCfg.ActiveDeadlineSeconds > 0 {
		activeDeadlineSeconds = &podCfg.ActiveDeadlineSeconds
	}
	volumes := make([]v12.Volume, 0)
	if internal.IsApplyConfigFromEnvScriptFaulty(c.Spec.Broker.Image) {
		volumes = append(volumes, v12.Volume{
			Name: configEnvPyScriptVolume,
			VolumeSource: v12.VolumeSource{
				ConfigMap: &v12.ConfigMapVolumeSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: configMapName(c),
					},
					DefaultMode: &(ConfigMapDefaultMode),
				},
			},
		})
	}
	return v12.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: c.GetName(), // leave it to k8s for name uniqueness
			Labels:       podTemplateLabels,
			Namespace:    deploymentNamespace(c),
			Annotations:  internal.GenerateAnnotations(c.Spec.Broker.Annotations),
		},
		Spec: v12.PodSpec{
			Affinity:              &podCfg.Affinity,
			NodeSelector:          podCfg.NodeSelector,
			RestartPolicy:         podCfg.RestartPolicy,
			SecurityContext:       &podCfg.SecurityContext,
			ActiveDeadlineSeconds: activeDeadlineSeconds,
			Containers:            []v12.Container{createDeploymentPodContainer(c)},
			Volumes:               volumes,
		},
	}
}

func createDeploymentPodContainer(c *v1alpha1.PulsarCluster) v12.Container {
	volumeMounts := make([]v12.VolumeMount, 0)
	applyConfigFromEnvScriptDirectory := "bin"
	if internal.IsApplyConfigFromEnvScriptFaulty(c.Spec.Broker.Image) {
		applyConfigFromEnvScriptDirectory = "/config"
		volumeMounts = append(volumeMounts, v12.VolumeMount{
			Name:      configEnvPyScriptVolume,
			MountPath: applyConfigFromEnvScriptDirectory,
			ReadOnly:  false,
		})
	}
	return v12.Container{
		Name:            "pulsar-broker",
		VolumeMounts:    volumeMounts,
		Command:         k8s.ContainerShellCommand(),
		Image:           c.Spec.Broker.Image.Name(),
		ImagePullPolicy: c.Spec.Broker.Image.PullPolicy,
		Resources:       c.Spec.Broker.PodConfig.Resources,
		Ports:           createDeploymentPodContainerPorts(c),
		Args:            createDeploymentPodContainerArguments(applyConfigFromEnvScriptDirectory),
		EnvFrom:         c.Spec.Broker.PodConfig.GenerateEnvFrom(),
		Env:             c.Spec.Broker.PodConfig.GenerateEnvVar(brokerConfigs(c)...),
	}
}

func createDeploymentPodContainerArguments(applyConfigScriptDirectory string) []string {
	return []string{
		fmt.Sprintf("%s/apply-config-from-env.py conf/broker.conf && bin/pulsar broker",
			applyConfigScriptDirectory),
	}
}

func createDeploymentPodContainerPorts(_ *v1alpha1.PulsarCluster) []v12.ContainerPort {
	return []v12.ContainerPort{
		{
			Name:          "pulsar-tcp",
			Protocol:      v12.ProtocolTCP,
			ContainerPort: internal.ServicePort,
		},
		{
			Name:          "pulsar-tls",
			Protocol:      v12.ProtocolTCP,
			ContainerPort: internal.ServicePortTLS,
		},
		{
			Name:          "pulsar-http",
			Protocol:      v12.ProtocolTCP,
			ContainerPort: internal.WebServicePort,
		},
		{
			Name:          "pulsar-https",
			Protocol:      v12.ProtocolTCP,
			ContainerPort: internal.WebServicePortTLS,
		},
	}
}

func brokerConfigs(c *v1alpha1.PulsarCluster) []v12.EnvVar {
	envVars := make([]v12.EnvVar, 0)
	addToEnv := func(source map[string]string) {
		for k, v := range source {
			envVars = append(envVars, v12.EnvVar{
				Name:  pulsarConfigEnvPrefix + k,
				Value: v,
			})
		}
	}
	if c.Spec.Broker.Configs != "" {
		brokerConf := map[string]string{}
		if err := yaml.Unmarshal([]byte(c.Spec.Broker.Configs), brokerConf); err != nil {
			fmt.Println(fmt.Errorf("invalid broker.conf data. reason: %s", err))
		}
		addToEnv(brokerConf)
	}
	envVars = replacePulsarConfigEnv(envVars, "clusterName", c.GetName())
	envVars = replacePulsarConfigEnv(envVars, "zookeeperServers", c.Spec.Broker.ZookeeperServers)
	envVars = replacePulsarConfigEnv(envVars, "configurationStoreServers", c.Spec.Broker.ConfigurationStoreServers)
	return envVars
}

func replacePulsarConfigEnv(envVars []v12.EnvVar, name string, value string) []v12.EnvVar {
	hasEnv := false
	newEnv := v12.EnvVar{Name: pulsarConfigEnvPrefix + name, Value: value}
	for i, env := range envVars {
		if hasEnv = env.Name == newEnv.Name; hasEnv {
			envVars[i] = newEnv
			break
		}
	}
	if !hasEnv {
		envVars = append(envVars, newEnv)
	}
	return envVars
}
