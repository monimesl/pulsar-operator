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
	"github.com/skulup/operator-pkg/reconciler"
	"github.com/skulup/pulsar-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	_              reconciler.Context    = &Reconciler{}
	_              reconciler.Reconciler = &Reconciler{}
	reconcileFuncs                       = []func(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error{
		reconcileClusterMetadata,
		reconcileConfigMap,
		reconcileDeployment,
		reconcileServices,
		reconcileClusterStage,
	}
)

// Reconciler reconciles a PulsarCluster object
type Reconciler struct {
	reconciler.Context
}

// Configure configures the reconciler
func (r *Reconciler) Configure(ctx reconciler.Context) error {
	r.Context = ctx
	return ctx.NewControllerBuilder().
		For(&v1alpha1.PulsarCluster{}).
		Owns(&v1beta1.PodDisruptionBudget{}).
		Owns(&v1.Deployment{}).
		Owns(&v12.ConfigMap{}).
		Owns(&v12.Service{}).
		Owns(&v13.Job{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=pulsar.skulup.com,resources=pulsarclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pulsar.skulup.com,resources=pulsarclusters/status,verbs=get;update;patch
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	cluster := &v1alpha1.PulsarCluster{}
	return r.Run(request, cluster, func() error {
		for _, fun := range reconcileFuncs {
			if err := fun(r.Context, cluster); err != nil {
				return err
			}
		}
		return nil
	})
}
