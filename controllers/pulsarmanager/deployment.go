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

package pulsarmanager

import (
	"context"
	"fmt"
	"github.com/skulup/operator-helper/k8s/deployment"
	"github.com/skulup/operator-helper/reconciler"
	"github.com/skulup/pulsar-operator/api/v1alpha1"
	"github.com/skulup/pulsar-operator/internal"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func reconcileDeployment(ctx reconciler.Context, manager *v1alpha1.PulsarManager) error {
	dep := &v1.Deployment{}
	return ctx.GetResource(types.NamespacedName{
		Namespace: deploymentNamespace(manager),
		Name:      deploymentName(manager),
	}, dep,
		func() (err error) { // Deployment Found
			return err
		},
		func() (err error) { // Deployment not Found
			dep = createDeployment(manager)
			if err = ctx.SetOwnershipReference(manager, dep); err == nil {
				if err = ctx.Client().Create(context.TODO(), dep); err == nil {
					ctx.Logger().Info("Deployment creation success. ",
						"Deployment.Name", dep.GetName(),
						"Deployment.Namespace", dep.GetNamespace())
				}
			}
			return
		})
}

func deploymentNamespace(manager *v1alpha1.PulsarManager) string {
	return manager.GetNamespace()
}

func deploymentName(manager *v1alpha1.PulsarManager) string {
	return fmt.Sprintf("%s-pulsar-manager", manager.GetName())
}

func volumeName(manager *v1alpha1.PulsarManager) string {
	return fmt.Sprintf("%s-pulsar-manager", manager.GetName())
}

func createDeployment(manager *v1alpha1.PulsarManager) *v1.Deployment {
	Replicas := int32(1)
	labels := internal.GenerateLabels(internal.Manager, manager.Spec.GeneratePodLabels())
	return deployment.New(deploymentNamespace(manager), deploymentName(manager), manager.Spec.Labels,
		v1.DeploymentSpec{
			Replicas: &Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: createDeploymentPodTemplateSpec(manager, labels),
		})
}

func createDeploymentPodTemplateSpec(manager *v1alpha1.PulsarManager, labels map[string]string) v12.PodTemplateSpec {
	podCfg := manager.Spec.PodConfig
	var activeDeadlineSeconds *int64
	if podCfg.ActiveDeadlineSeconds > 0 {
		activeDeadlineSeconds = &podCfg.ActiveDeadlineSeconds
	}
	return v12.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: manager.GetName(), // leave it to k8s for name uniqueness
			Namespace:    deploymentNamespace(manager),
			Labels:       labels,
			Annotations:  internal.GenerateAnnotations(manager.Spec.Annotations),
		},
		Spec: v12.PodSpec{
			Affinity:              &podCfg.Affinity,
			NodeSelector:          podCfg.NodeSelector,
			RestartPolicy:         podCfg.RestartPolicy,
			SecurityContext:       &podCfg.SecurityContext,
			ActiveDeadlineSeconds: activeDeadlineSeconds,
			Containers:            []v12.Container{createDeploymentPodContainer(manager)},
			Volumes:               []v12.Volume{createDeploymentPodTemplateSpecVolume(manager)},
		},
	}
}

func createDeploymentPodTemplateSpecVolume(manager *v1alpha1.PulsarManager) v12.Volume {
	return v12.Volume{
		Name:         volumeName(manager),
		VolumeSource: v12.VolumeSource{EmptyDir: &v12.EmptyDirVolumeSource{}},
	}
}

func createDeploymentPodContainer(manager *v1alpha1.PulsarManager) v12.Container {
	return v12.Container{
		Name:            "pulsar-manager",
		Image:           manager.Spec.Image.Name(),
		ImagePullPolicy: manager.Spec.Image.PullPolicy,
		Resources:       manager.Spec.PodConfig.Resources,
		Ports:           createDeploymentPodContainerPorts(),
		VolumeMounts:    createDeploymentPodContainerVolumeMount(manager),
		Env:             manager.Spec.PodConfig.GenerateEnvVar(createDeploymentPodContainerEnvs(manager)...),
	}
}

func createDeploymentPodContainerVolumeMount(manager *v1alpha1.PulsarManager) []v12.VolumeMount {
	return []v12.VolumeMount{
		{
			Name:      volumeName(manager),
			MountPath: v1alpha1.ManagerDefaultVolumeMountPath,
		},
	}
}

func createDeploymentPodContainerPorts() []v12.ContainerPort {
	return []v12.ContainerPort{
		{
			Name:          "backend",
			Protocol:      v12.ProtocolTCP,
			ContainerPort: internal.ManagerBackendPort,
		},
		{
			Name:          "frontend",
			Protocol:      v12.ProtocolTCP,
			ContainerPort: internal.ManagerFrontendPort,
		},
	}
}

func createDeploymentPodContainerEnvs(c *v1alpha1.PulsarManager) []v12.EnvVar {
	envs := []v12.EnvVar{
		{
			Name:  "REDIRECT_HOST",
			Value: "127.0.0.1",
		},
		{
			Name:  "REDIRECT_PORT",
			Value: "9527",
		},
		{
			Name:  "DRIVER_CLASS_NAME",
			Value: "org.postgresql.Driver",
		},
		{
			Name:  "URL",
			Value: "jdbc:postgresql://127.0.0.1:5432/pulsar_manager",
		},
		{
			Name:  "USERNAME",
			Value: "pulsar",
		},
		{
			Name:  "PASSWORD",
			Value: c.Spec.DbPassword,
		},
		{
			Name:  "LOG_LEVEL",
			Value: c.Spec.LogLevel,
		},
	}
	return envs
}
