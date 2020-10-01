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
	"github.com/wireltd/operator-pkg/reconciler"
	"github.com/wireltd/pulsar-operator/api/v1alpha1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	_              reconciler.Context    = &Reconciler{}
	_              reconciler.Reconciler = &Reconciler{}
	reconcileFuncs                       = []func(ctx reconciler.Context, cluster *v1alpha1.PulsarManager) error{
		reconcileDeployment,
		reconcileService,
		reconcileSecretes,
		reconcileSuperUserAccount,
	}
)

// Reconciler reconciles a PulsarManager object
type Reconciler struct {
	reconciler.Context
}

// Configure configures the reconciler
func (r Reconciler) Configure(ctx reconciler.Context) error {
	r.Context = ctx
	return ctx.NewControllerBuilder().
		For(&v1alpha1.PulsarManager{}).
		Owns(&appsV1.Deployment{}).
		Owns(&coreV1.ConfigMap{}).
		Owns(&coreV1.Service{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=pulsar.wirelimited.com,resources=pulsarmanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pulsar.wirelimited.com,resources=pulsarmanagers/status,verbs=get;update;patch

// Reconciler performs a full reconciliation for the object referred to by the Request.
func (r Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	manager := &v1alpha1.PulsarManager{}
	return r.Run(request, manager, func() error {
		for _, fun := range reconcileFuncs {
			if err := fun(r.Context, manager); err != nil {
				return err
			}
		}
		return nil
	})
}
