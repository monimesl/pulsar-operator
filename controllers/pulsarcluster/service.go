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
	"github.com/alphashaw/operator-pkg/k8s/service"
	"github.com/alphashaw/operator-pkg/reconciler"
	"github.com/alphashaw/pulsar-operator/api/v1alpha1"
	"github.com/alphashaw/pulsar-operator/internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func reconcileServices(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) (err error) {
	if err = reconcileClusterIpService(ctx, cluster); err == nil {
		err = reconcileHeadlessService(ctx, cluster)
	}
	return
}

func reconcileClusterIpService(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	svcName := clusterIpServiceName(cluster)
	svc := createService(cluster, svcName, "")
	return reconcileService(ctx, cluster, svc)
}

func reconcileHeadlessService(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	svcName := headlessServiceName(cluster)
	svc := createService(cluster, svcName, "None")
	return reconcileService(ctx, cluster, svc)
}

func reconcileService(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster, newSvc *v1.Service) error {
	if cluster.IsInitialized() {
		currentSvc := &v1.Service{}
		return ctx.GetResource(types.NamespacedName{
			Namespace: newSvc.Namespace,
			Name:      newSvc.Name,
		}, currentSvc,
			func() (err error) {
				return // Service Found
			},
			func() (err error) {
				// Service Not Found
				if err = ctx.SetOwnershipReference(cluster, newSvc); err == nil {
					if err = ctx.Client().Create(context.TODO(), newSvc); err == nil {
						ctx.Logger().Info("Service creation success. ",
							"Service.Name", newSvc.GetName(),
							"Service.Namespace", newSvc.GetNamespace())
					}
				}
				return
			})
	}
	return nil
}

func createService(c *v1alpha1.PulsarCluster, name string, clusterIp string) *v1.Service {
	labels := internal.GenerateLabels(internal.Broker, c.Spec.Broker.GeneratePodLabels(c.GetName()))
	return service.New(serviceNamespace(c), name, labels, v1.ServiceSpec{
		Type:      v1.ServiceTypeClusterIP,
		ClusterIP: clusterIp,
		Selector:  labels,
		Ports:     internal.CreateServicePorts(),
	})
}

func serviceNamespace(c *v1alpha1.PulsarCluster) string {
	return c.GetNamespace()
}

func clusterIpServiceName(c *v1alpha1.PulsarCluster) string {
	return fmt.Sprintf("%s-pulsar-broker", c.GetName())
}

func headlessServiceName(c *v1alpha1.PulsarCluster) string {
	return fmt.Sprintf("%s-pulsar-broker-headless", c.GetName())
}
