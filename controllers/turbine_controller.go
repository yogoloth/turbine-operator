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
	"time"
	//"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitorwangjldevv1beta1 "github.com/yogoloth/turbine-operator/api/v1beta1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TurbineReconciler reconciles a Turbine object
type TurbineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	serviceOwnerKey string = ".metadata.controller"
)

func constructServiceForDashboard(turbine *monitorwangjldevv1beta1.Turbine, schema *runtime.Scheme) (*corev1.Service, error) {

	port := corev1.ServicePort{Name: "dashboard", Port: 9002}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        "turbinedashboard-" + turbine.Name,
			Namespace:   turbine.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"turbine-monitor": turbine.Name},
			Type:     corev1.ServiceTypeNodePort,
			Ports:    []corev1.ServicePort{port},
		},
	}
	svc.Labels["runner"] = "turbineoperator"
	svc.Labels["turbine-monitor-dashboard"] = turbine.Name
	//svc.Labels[""] = monitorwangjldevv1beta1.MonitorName

	if err := ctrl.SetControllerReference(turbine, svc, schema); err != nil {
		return nil, err
	}

	return svc, nil
}

func constructServiceForTurbine(hystrix *monitorwangjldevv1beta1.Hystrix, turbine *monitorwangjldevv1beta1.Turbine, schema *runtime.Scheme) (*corev1.Service, error) {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        "turbineoperator-" + turbine.Name + "-" + hystrix.Name,
			Namespace:   turbine.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: hystrix.Selector,
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    hystrix.Ports,
		},
	}
	svc.Labels["runner"] = "turbineoperator"
	svc.Labels["turbine-monitor"] = turbine.Name
	//svc.Labels[""] = monitorwangjldevv1beta1.MonitorName

	if err := ctrl.SetControllerReference(turbine, svc, schema); err != nil {
		return nil, err
	}

	return svc, nil
}

// +kubebuilder:rbac:groups=monitor.wangjl.dev.wangjl.dev,resources=turbines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitor.wangjl.dev.wangjl.dev,resources=turbines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch

func (r *TurbineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("turbine", req.NamespacedName)
	scheduledResult := ctrl.Result{RequeueAfter: 10 * time.Second}

	turbineMonitor := monitorwangjldevv1beta1.Turbine{}
	if err := r.Get(ctx, req.NamespacedName, &turbineMonitor); err != nil {
		logger.Error(err, "get turbineMonitor")
		return scheduledResult, err
	}

	tmp_svc := corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: turbineMonitor.Namespace, Name: "turbinedashboard-" + turbineMonitor.Name}, &tmp_svc); err != nil {
		if svc, err := constructServiceForDashboard(&turbineMonitor, r.Scheme); err != nil {
			logger.Error(err, "constructServiceForDashboard")
			return scheduledResult, err
		} else if err := r.Create(ctx, svc); err != nil {
			logger.Error(err, "create ServiceForDashboard", "svc", svc)
			return scheduledResult, err
		}
	}

	//if monitorwangjldevv1beta1.MonitorName == "" {
	//	monitorwangjldevv1beta1.MonitorName = turbineMonitor.Name
	//}

	for _, hystrix := range turbineMonitor.Spec.Hystrixs {
		if svc, err := constructServiceForTurbine(&hystrix, &turbineMonitor, r.Scheme); err != nil {
			logger.Error(err, "constructServiceForTurbine")
			return scheduledResult, err
		} else {
			tmp_svc := corev1.Service{}
			if err := r.Get(ctx, types.NamespacedName{Namespace: turbineMonitor.Namespace, Name: "turbineoperator-" + turbineMonitor.Name + "-" + hystrix.Name}, &tmp_svc); err != nil {
				logger.Info("createService")
				if err := r.Create(ctx, svc); err != nil {
					logger.Error(err, "createService")
					return scheduledResult, err
				}
			} else {
				logger.V(0).Info("service exists", "service", tmp_svc)
			}
		}
	}

	//var HystrixList corev1.ServiceList
	//if err := r.List(ctx, &monitorwangjldevv1beta1.HystrixList, client.InNamespace(req.Namespace), client.MatchingFields{serviceOwnerKey: req.Name}); err != nil {
	//	logger.Error(err, "set hystrixlist")
	//}

	//logger.Info("set hystrixlist", "hystrixlist", monitorwangjldevv1beta1.HystrixList)

	return ctrl.Result{}, nil
}

func (r *TurbineReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(&corev1.Service{}, serviceOwnerKey, func(rawObj runtime.Object) []string {
		service := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(service)
		if owner == nil {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&monitorwangjldevv1beta1.Turbine{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
