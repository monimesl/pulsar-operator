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
	"github.com/monimesl/operator-helper/k8s/configmap"
	"github.com/monimesl/operator-helper/oputil"
	"github.com/monimesl/operator-helper/reconciler"
	"github.com/monimesl/pulsar-operator/api/v1alpha1"
	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

// ReconcileConfigMap reconcile the configmap of the specified cluster
func ReconcileConfigMap(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	cm := &v1.ConfigMap{}
	return ctx.GetResource(types.NamespacedName{
		Name:      cluster.ConfigMapName(),
		Namespace: cluster.Namespace,
	}, cm,
		nil,
		// Not Found
		func() (err error) {
			cm = createConfigMap(cluster)
			if err = ctx.SetOwnershipReference(cluster, cm); err == nil {
				ctx.Logger().Info("Creating the pulsar configMap",
					"ConfigMap.Name", cm.GetName(),
					"ConfigMap.Namespace", cm.GetNamespace())
				if err = ctx.Client().Create(context.TODO(), cm); err == nil {
					ctx.Logger().Info("ConfigMap creation success.",
						"ConfigMap.Name", cm.GetName(),
						"ConfigMap.Namespace", cm.GetNamespace())
				}
			}
			return
		})
}

func createConfigMap(cluster *v1alpha1.PulsarCluster) *v1.ConfigMap {
	jvmOptions := cluster.Spec.JVMOptions
	excludedOptions := []string{
		"PULSAR_GC", "PULSAR_MEM", "PULSAR_EXTRA_OPTS", "PULSAR_GC_LOG",
	}
	data := map[string]string{
		"PULSAR_GC":         strings.Join(jvmOptions.Gc, " "),
		"PULSAR_MEM":        strings.Join(jvmOptions.Memory, " "),
		"PULSAR_EXTRA_OPTS": strings.Join(jvmOptions.Extra, " "),
		"PULSAR_GC_LOG":     strings.Join(jvmOptions.GcLogging, " "),
	}
	for k, v := range cluster.Spec.BrokerConfig {
		if !strings.HasPrefix(k, "PULSAR_PREFIX_") {
			k = fmt.Sprintf("PULSAR_PREFIX_%s", k)
		}
		if oputil.Contains(excludedOptions, k) {
			log.Warnf("ignoring the config: %s", k)
			continue
		}
		data[k] = v
	}
	return configmap.New(cluster.Namespace, cluster.ConfigMapName(), data)
}
