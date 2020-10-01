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
	"github.com/skulup/operator-helper/k8s/service"
	"github.com/skulup/operator-helper/reconciler"
	"github.com/skulup/pulsar-operator/api/v1alpha1"
	"github.com/skulup/pulsar-operator/internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func reconcileService(ctx reconciler.Context, manager *v1alpha1.PulsarManager) error {
	svc := &v1.Service{}
	return ctx.GetResource(types.NamespacedName{
		Namespace: serviceNamespace(manager),
		Name:      serviceName(manager),
	}, svc,
		func() (err error) {
			return // Service Found
		},
		func() (err error) {
			// Service Not Found
			svc = createService(manager)
			if err = ctx.SetOwnershipReference(manager, svc); err == nil {
				if err = ctx.Client().Create(context.TODO(), svc); err == nil {
					ctx.Logger().Info("Service creation success. ",
						"Service.Name", svc.GetName(),
						"Service.Namespace", svc.GetNamespace())
				}
			}
			return
		})
}

func createService(manager *v1alpha1.PulsarManager) *v1.Service {
	labels := internal.GenerateLabels(internal.Manager, manager.Spec.GeneratePodLabels())
	return service.New(serviceNamespace(manager), serviceName(manager), labels, v1.ServiceSpec{
		Type:     v1.ServiceTypeClusterIP,
		Selector: labels,
		Ports: []v1.ServicePort{
			{
				Name:     "backend",
				Protocol: v1.ProtocolTCP,
				Port:     internal.ManagerBackendPort,
			},
			{
				Name:     "frontend",
				Protocol: v1.ProtocolTCP,
				Port:     internal.ManagerFrontendPort,
			},
		},
	})
}

func serviceNamespace(manager *v1alpha1.PulsarManager) string {
	return manager.GetNamespace()
}

func serviceName(c *v1alpha1.PulsarManager) string {
	return fmt.Sprintf("%s-pulsar-manager", c.GetName())
}
