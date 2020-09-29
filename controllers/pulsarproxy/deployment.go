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

package pulsarproxy

import (
	"context"
	"fmt"
	"github.com/skulup/operator-pkg/k8s"
	"github.com/skulup/operator-pkg/k8s/deployment"
	"github.com/skulup/operator-pkg/reconciler"
	"github.com/skulup/pulsar-operator/api/v1alpha1"
	"github.com/skulup/pulsar-operator/internal"
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

func reconcileDeployment(ctx reconciler.Context, proxy *v1alpha1.PulsarProxy) error {
	dep := &v1.Deployment{}
	return ctx.GetResource(types.NamespacedName{
		Namespace: deploymentNamespace(proxy),
		Name:      deploymentName(proxy),
	}, dep,
		func() (err error) { // Deployment Found
			updated := false
			oldReplicaCount := *dep.Spec.Replicas
			if oldReplicaCount != proxy.Spec.Count {
				updated = true
				dep.Spec.Replicas = &proxy.Spec.Count
			}
			if updated {
				if err = ctx.Client().Update(context.TODO(), dep); err == nil {
					ctx.Logger().Info("Deployment replica set scale successfully. ",
						"Deployment.Name", dep.GetName(),
						"Deployment.Namespace", dep.GetNamespace(),
						"Old Deployment.ReplicaSize", oldReplicaCount,
						"New Deployment.ReplicaSize", proxy.Spec.Count)
				}
			}
			return err
		},
		func() (err error) { // Deployment not Found
			dep = createDeployment(proxy)
			if err = ctx.SetOwnershipReference(proxy, dep); err == nil {
				if err = ctx.Client().Create(context.TODO(), dep); err == nil {
					ctx.Logger().Info("Deployment creation success. ",
						"Deployment.Name", dep.GetName(),
						"Deployment.Namespace", dep.GetNamespace())
				}
			}
			return
		})
}

func deploymentNamespace(proxy *v1alpha1.PulsarProxy) string {
	return proxy.GetNamespace()
}

func deploymentName(proxy *v1alpha1.PulsarProxy) string {
	return fmt.Sprintf("%s-pulsar-proxy", proxy.GetName())
}

func createDeployment(proxy *v1alpha1.PulsarProxy) *v1.Deployment {
	labels := internal.GenerateLabels(internal.Proxy, proxy.Spec.Proxy.GeneratePodLabels())
	return deployment.New(deploymentNamespace(proxy), deploymentName(proxy), proxy.Spec.Proxy.Labels,
		v1.DeploymentSpec{
			Replicas: &proxy.Spec.Count,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: createDeploymentPodTemplateSpec(proxy, labels),
		})
}

func createDeploymentPodTemplateSpec(proxy *v1alpha1.PulsarProxy, labels map[string]string) v12.PodTemplateSpec {
	DefMode := int32(0744)
	podCfg := proxy.Spec.Proxy.PodConfig
	var activeDeadlineSeconds *int64
	if podCfg.ActiveDeadlineSeconds > 0 {
		activeDeadlineSeconds = &podCfg.ActiveDeadlineSeconds
	}
	volumes := make([]v12.Volume, 0)
	if internal.IsApplyConfigFromEnvScriptFaulty(proxy.Spec.Proxy.Image) {
		volumes = append(volumes, v12.Volume{
			Name: configEnvPyScriptVolume,
			VolumeSource: v12.VolumeSource{
				ConfigMap: &v12.ConfigMapVolumeSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: configMapName(proxy),
					},
					DefaultMode: &(DefMode),
				},
			},
		})
	}
	return v12.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: proxy.GetName(), // leave it to k8s for name uniqueness
			Namespace:    deploymentNamespace(proxy),
			Labels:       labels,
			Annotations:  internal.GenerateAnnotations(proxy.Spec.Proxy.Annotations),
		},
		Spec: v12.PodSpec{
			Affinity:              &podCfg.Affinity,
			NodeSelector:          podCfg.NodeSelector,
			RestartPolicy:         podCfg.RestartPolicy,
			SecurityContext:       &podCfg.SecurityContext,
			ActiveDeadlineSeconds: activeDeadlineSeconds,
			Containers:            []v12.Container{createDeploymentContainer(proxy)},
			Volumes:               volumes,
		},
	}
}

func createDeploymentContainer(proxy *v1alpha1.PulsarProxy) v12.Container {
	volumeMounts := make([]v12.VolumeMount, 0)
	applyConfigFromEnvScriptDirectory := "bin"
	if internal.IsApplyConfigFromEnvScriptFaulty(proxy.Spec.Proxy.Image) {
		applyConfigFromEnvScriptDirectory = "/config"
		volumeMounts = append(volumeMounts, v12.VolumeMount{
			Name:      configEnvPyScriptVolume,
			MountPath: applyConfigFromEnvScriptDirectory,
			ReadOnly:  false,
		})
	}
	return v12.Container{
		Name:            "pulsar-proxy",
		VolumeMounts:    volumeMounts,
		Command:         k8s.ContainerShellCommand(),
		Image:           proxy.Spec.Proxy.Image.Name(),
		ImagePullPolicy: proxy.Spec.Proxy.Image.PullPolicy,
		Resources:       proxy.Spec.Proxy.PodConfig.Resources,
		Ports:           createDeploymentContainerPorts(proxy),
		EnvFrom:         proxy.Spec.Proxy.PodConfig.GenerateEnvFrom(),
		Env:             proxy.Spec.Proxy.PodConfig.GenerateEnvVar(proxyConfigs(proxy)...),
		Args:            createDeploymentContainerArguments(applyConfigFromEnvScriptDirectory),
	}
}

func createDeploymentContainerArguments(applyConfigScriptDirectory string) []string {
	return []string{
		fmt.Sprintf("%s/apply-config-from-env.py conf/proxy.conf && bin/pulsar proxy",
			applyConfigScriptDirectory),
	}
}

func createDeploymentContainerPorts(_ *v1alpha1.PulsarProxy) []v12.ContainerPort {
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

func proxyConfigs(proxy *v1alpha1.PulsarProxy) []v12.EnvVar {
	envVars := make([]v12.EnvVar, 0)
	addToEnv := func(source map[string]string) {
		for k, v := range source {
			envVars = append(envVars, v12.EnvVar{
				Name:  pulsarConfigEnvPrefix + k,
				Value: v,
			})
		}
	}
	if proxy.Spec.Proxy.Configs != "" {
		proxyConf := map[string]string{}
		if err := yaml.Unmarshal([]byte(proxy.Spec.Proxy.Configs), proxyConf); err != nil {
			fmt.Println(fmt.Errorf("invalid proxy.conf data. reason: %s", err))
		}
		addToEnv(proxyConf)
	}
	return envVars
}
