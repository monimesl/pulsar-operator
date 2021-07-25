/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/monimesl/operator-helper/reconciler"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pulsarv1alpha1 "github.com/monimesl/pulsar-operator/api/v1alpha1"
)

var (
	_                   reconciler.Context    = &PulsarProxyReconciler{}
	_                   reconciler.Reconciler = &PulsarProxyReconciler{}
	proxyReconcileFuncs                       = []func(ctx reconciler.Context, cluster *pulsarv1alpha1.PulsarProxy) error{}
)

// PulsarProxyReconciler reconciles a PulsarProxy object
type PulsarProxyReconciler struct {
	reconciler.Context
}

// Configure configures the above PulsarProxyReconciler
func (r *PulsarProxyReconciler) Configure(ctx reconciler.Context) error {
	r.Context = ctx
	return ctx.NewControllerBuilder().
		For(&pulsarv1alpha1.PulsarProxy{}).
		Owns(&v1beta1.PodDisruptionBudget{}).
		Owns(&v12.StatefulSet{}).
		Owns(&v1.ConfigMap{}).
		Owns(&v1.Service{}).
		Complete(r)
}

// Reconcile handles reconciliation request for ZookeeperCluster instances
func (r *PulsarProxyReconciler) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	cluster := &pulsarv1alpha1.PulsarProxy{}
	return r.Run(request, cluster, func(_ bool) (err error) {
		for _, fun := range proxyReconcileFuncs {
			if err = fun(r, cluster); err != nil {
				break
			}
		}
		return
	})
}
