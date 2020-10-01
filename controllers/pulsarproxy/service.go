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
	"github.com/skulup/operator-helper/k8s/service"
	"github.com/skulup/operator-helper/reconciler"
	"github.com/skulup/pulsar-operator/api/v1alpha1"
	"github.com/skulup/pulsar-operator/internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func reconcileServices(ctx reconciler.Context, proxy *v1alpha1.PulsarProxy) (err error) {
	if err = reconcileClusterIpService(ctx, proxy); err == nil {
		err = reconcileHeadlessService(ctx, proxy)
	}
	return
}

func reconcileClusterIpService(ctx reconciler.Context, proxy *v1alpha1.PulsarProxy) error {
	svcName := clusterIpServiceName(proxy)
	svc := createService(proxy, svcName, "")
	return reconcileService(ctx, proxy, svc)
}

func reconcileHeadlessService(ctx reconciler.Context, proxy *v1alpha1.PulsarProxy) error {
	svcName := headlessServiceName(proxy)
	svc := createService(proxy, svcName, "None")
	return reconcileService(ctx, proxy, svc)
}

func reconcileService(ctx reconciler.Context, proxy *v1alpha1.PulsarProxy, newSvc *v1.Service) error {
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
			if err = ctx.SetOwnershipReference(proxy, newSvc); err == nil {
				if err = ctx.Client().Create(context.TODO(), newSvc); err == nil {
					ctx.Logger().Info("Service creation success. ",
						"Service.Name", newSvc.GetName(),
						"Service.Namespace", newSvc.GetNamespace())
				}
			}
			return
		})
}

func createService(proxy *v1alpha1.PulsarProxy, name string, clusterIp string) *v1.Service {
	labels := internal.GenerateLabels(internal.Proxy, proxy.Spec.Proxy.GeneratePodLabels())
	return service.New(serviceNamespace(proxy), name, labels, v1.ServiceSpec{
		Type:      v1.ServiceTypeClusterIP,
		ClusterIP: clusterIp,
		Selector:  labels,
		Ports:     internal.CreateServicePorts(),
	})
}

func serviceNamespace(proxy *v1alpha1.PulsarProxy) string {
	return proxy.GetNamespace()
}

func clusterIpServiceName(proxy *v1alpha1.PulsarProxy) string {
	return fmt.Sprintf("%s-pulsar-proxy", proxy.GetName())
}

func headlessServiceName(proxy *v1alpha1.PulsarProxy) string {
	return fmt.Sprintf("%s-pulsar-proxy-headless", proxy.GetName())
}
