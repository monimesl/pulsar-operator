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
	"github.com/alphashaw/operator-pkg/reconciler"
	"github.com/alphashaw/pulsar-operator/api/v1alpha1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	_              reconciler.Context    = &Reconciler{}
	_              reconciler.Reconciler = &Reconciler{}
	reconcileFuncs                       = []func(ctx reconciler.Context, proxy *v1alpha1.PulsarProxy) error{
		reconcileConfigMap,
		reconcileDeployment,
		reconcileServices,
	}
)

// Reconciler reconciles a PulsarProxy object
type Reconciler struct {
	reconciler.Context
}

// Configure configures the reconciler
func (r *Reconciler) Configure(ctx reconciler.Context) error {
	r.Context = ctx
	return ctx.NewControllerBuilder().
		For(&v1alpha1.PulsarProxy{}).
		Owns(&appsV1.Deployment{}).
		Owns(&coreV1.ConfigMap{}).
		Owns(&coreV1.Service{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=pulsar.skulup.com,resources=pulsarproxies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pulsar.skulup.com,resources=pulsarproxies/status,verbs=get;update;patch

// Reconciler performs a full reconciliation for the object referred to by the Request.
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	proxy := &v1alpha1.PulsarProxy{}
	return r.Run(request, proxy, func() error {
		for _, fun := range reconcileFuncs {
			if err := fun(r.Context, proxy); err != nil {
				return err
			}
		}
		return nil
	})
}
