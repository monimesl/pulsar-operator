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
	"github.com/skulup/operator-pkg/k8s"
	"github.com/skulup/operator-pkg/k8s/job"
	"github.com/skulup/operator-pkg/reconciler"
	"github.com/skulup/pulsar-operator/api/v1alpha1"
	"github.com/skulup/pulsar-operator/pkg"
	v1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func reconcileClusterMetadata(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	return initMetadata(ctx, cluster)
}

func reconcileClusterStage(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) (err error) {
	if cluster.Status.Stage == v1alpha1.ClusterStageLaunching && isDeploymentReady(ctx, cluster) {
		cluster.Status.Stage = v1alpha1.ClusterStageRunning
		err = ctx.Client().Status().Update(context.TODO(), cluster)
	}
	return
}

func initMetadata(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	if cluster.Status.Stage == v1alpha1.ClusterStageInitializing {
		jb := &v1.Job{}
		return ctx.GetResource(types.NamespacedName{
			Namespace: jobNamespace(cluster),
			Name:      jobName(cluster),
		}, jb,
			func() (err error) { // Job already exists
				if jb.Status.Succeeded > 0 {
					// Update the PulsarCluster Status Stage to Launching
					cluster.Status.Stage = v1alpha1.ClusterStageLaunching
					if err = ctx.Client().Status().Update(context.TODO(), cluster); err == nil {
						ctx.Logger().Info("Pulsar cluster metadata initialization successful. ",
							"Job.Name", jb.GetName(),
							"Job.Namespace", jb.GetNamespace())
					}
				}
				return err
			},
			func() (err error) { // Job does not exists
				jb = createJob(cluster)
				if err = ctx.SetOwnershipReference(cluster, jb); err == nil {
					if err = ctx.Client().Create(context.TODO(), jb); err == nil {
						ctx.Logger().Info("Pulsar cluster metadata init job created successfully ",
							"Job.Name", jb.GetName(),
							"Job.Namespace", jb.GetNamespace())
					}
				}
				return err
			})
	}
	return nil
}

func createJob(c *v1alpha1.PulsarCluster) *v1.Job {
	labels := pkg.GenerateLabels("", nil)
	return job.New(jobNamespace(c), jobName(c), labels,
		v1.JobSpec{
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: jobName(c),
					Namespace:    jobNamespace(c),
					Labels:       labels,
				},
				Spec: createJobPodSpec(c),
			},
		})
}

func createJobPodSpec(c *v1alpha1.PulsarCluster) coreV1.PodSpec {
	return coreV1.PodSpec{
		RestartPolicy: coreV1.RestartPolicyNever,
		Containers:    []coreV1.Container{createJobPodContainer(c)},
	}
}

func createJobPodContainer(c *v1alpha1.PulsarCluster) coreV1.Container {
	return coreV1.Container{
		Name:            "cluster-metadata-init",
		Image:           c.Spec.Broker.Image.Name(),
		ImagePullPolicy: c.Spec.Broker.Image.PullPolicy,
		Command:         k8s.ContainerShellCommand(),
		Args:            createJobPodContainerArguments(c),
	}
}

func createJobPodContainerArguments(c *v1alpha1.PulsarCluster) []string {
	brokerFQDNService := fmt.Sprintf("%s.%s.svc.cluster.local", clusterIpServiceName(c), serviceNamespace(c))
	return []string{
		"bin/pulsar initialize-cluster-metadata " +
			fmt.Sprintf("--cluster %s ", c.GetName()) +
			fmt.Sprintf("--zookeeper %s ", c.Spec.Broker.ZookeeperServers) +
			fmt.Sprintf("--configuration-store %s ", c.Spec.Broker.ConfigurationStoreServers) +
			fmt.Sprintf("--web-service-url %s:%d ", brokerFQDNService, pkg.WebServicePort) +
			fmt.Sprintf("--web-service-url-tls %s:%d ", brokerFQDNService, pkg.WebServicePortTLS) +
			fmt.Sprintf("--broker-service-url %s:%d ", brokerFQDNService, pkg.ServicePort) +
			fmt.Sprintf("--broker-service-url-tls %s:%d ", brokerFQDNService, pkg.ServicePortTLS),
	}
}

func jobName(c *v1alpha1.PulsarCluster) string {
	return fmt.Sprintf("%s-metadata-init-job", c.GetName())
}

func jobNamespace(c *v1alpha1.PulsarCluster) string {
	return c.GetNamespace()
}
