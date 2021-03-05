/*
Copyright 2021 wangjl.

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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//monitorwangjldevv1beta1 "github.com/yogoloth/turbine-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// TurbineServerReconciler reconciles a TurbineServer object
type TurbineServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=monitor.wangjl.dev.wangjl.dev,resources=turbineservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitor.wangjl.dev.wangjl.dev,resources=turbineservers/status,verbs=get;update;patch

func (r *TurbineServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx = context.Background()
	logger = r.Log.WithValues("turbineserver", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *TurbineServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		//For(&monitorwangjldevv1beta1.TurbineServer{}).
		For(&corev1.Endpoints{}).
		Complete(r)
}
