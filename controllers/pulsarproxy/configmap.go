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
	"github.com/skulup/operator-pkg/reconciler"
	"github.com/skulup/pulsar-operator/api/v1alpha1"
	"github.com/skulup/pulsar-operator/internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func reconcileConfigMap(ctx reconciler.Context, proxy *v1alpha1.PulsarProxy) error {
	cm := &v1.ConfigMap{}
	return ctx.GetResource(types.NamespacedName{
		Namespace: configMapNamespace(proxy),
		Name:      configMapName(proxy),
	}, cm,
		func() (err error) { // Found
			return nil

		},
		func() (err error) {
			cm = internal.NewPulsarConfigMap(proxy.Spec.Proxy.Image,
				configMapNamespace(proxy), configMapName(proxy))
			if err = ctx.SetOwnershipReference(proxy, cm); err == nil {
				if err = ctx.Client().Create(context.TODO(), cm); err == nil {
					ctx.Logger().Info("ConfigMap creation success.",
						"ConfigMap.Name", cm.GetName(), "ConfigMap.Namespace", cm.GetNamespace())
				}
			}
			return
		})
}

func configMapNamespace(c *v1alpha1.PulsarProxy) string {
	return c.GetNamespace()
}

func configMapName(c *v1alpha1.PulsarProxy) string {
	return fmt.Sprintf("%s-pulsar-proxy", c.GetName())
}
