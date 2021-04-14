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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//"sigs.k8s.io/controller-runtime/pkg/log"

	monitorwangjldevv1beta1 "github.com/yogoloth/turbine-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TurbineServerReconciler reconciles a TurbineServer object
type TurbineServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type Hystrixs map[string]bool

//var (
//	turbines = map[string]hystrixs{}
//)

func constructTurbinePod(turbine *monitorwangjldevv1beta1.Turbine, schema *runtime.Scheme) (*corev1.Pod, error) {

	turbine_streams := corev1.EnvVar{Name: "TURBINE_STREAMS", Value: ""}

	turbine_container := corev1.Container{
		Name:  "hystrix-turbine",
		Image: "arthurtsang/docker-hystrix-turbine:latest",
		//TURBINE_STREAMS always be first env
		Env: []corev1.EnvVar{turbine_streams},
	}
	hystrix_dashboard_container := corev1.Container{
		Name:  "hystrix-dashboard",
		Image: "mlabouardy/hystrix-dashboard:latest",
	}
	hystrix_exporter_container := corev1.Container{
		Name:  "hystrix-metrics",
		Image: "wangjl/hystrix_exporter:v0.1_wangjl",
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        "turbineoperator-" + turbine.Name,
			Namespace:   turbine.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{turbine_container, hystrix_dashboard_container, hystrix_exporter_container},
		},
	}
	pod.Labels["runner"] = "turbineoperator"
	pod.Labels["turbine-monitor"] = turbine.Name

	pod.Annotations["turbine_streams"] = ""

	if err := ctrl.SetControllerReference(turbine, pod, schema); err != nil {
		return nil, err
	}

	return pod, nil
}

// +kubebuilder:rbac:groups=monitor.wangjl.dev.wangjl.dev,resources=turbines,verbs=get;list;watch
// +kubebuilder:rbac:groups=monitor.wangjl.dev.wangjl.dev,resources=turbines/status,verbs=get
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=endpoints/status,verbs=get

func (r *TurbineServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("turbineserver", req.NamespacedName)
	scheduledResult := ctrl.Result{RequeueAfter: 10 * time.Second}
	//turbines := map[string]hystrixs{}

	pod_exists := false

	my_hystrixs := Hystrixs{}

	if strings.HasPrefix(req.Name, "turbineoperator-") {

		ep := corev1.Endpoints{}
		if err := r.Get(ctx, req.NamespacedName, &ep); err != nil {
			return ctrl.Result{}, nil
		}

		turbine_name := ep.Labels["turbine-monitor"]

		if turbine_name != "" {
			old_pod := corev1.Pod{}
			pod_namespace_name := types.NamespacedName{Namespace: req.Namespace, Name: "turbineoperator-" + turbine_name}
			if err := r.Get(ctx, pod_namespace_name, &old_pod); err != nil {
				logger.Info("not found", "pod", pod_namespace_name)
			} else {
				logger.Info("found", "name", pod_namespace_name, "old_pod", old_pod)
				pod_exists = true
				turbine_streams := strings.Split(old_pod.Annotations["turbine_streams"], ",")
				for _, stream := range turbine_streams {
					my_hystrixs[stream] = true
				}
			}

			eps := corev1.EndpointsList{}
			if err := r.List(ctx, &eps, client.InNamespace(req.Namespace), client.MatchingLabels{"turbine-monitor": turbine_name}); err != nil {
				logger.Error(err, "unable to fetch podlist")
				return ctrl.Result{}, err
			}

			hs := Hystrixs{}
			for _, tmp_ep := range eps.Items {
				for _, subset := range tmp_ep.Subsets {
					for _, address := range subset.Addresses {
						for _, port := range subset.Ports {
							addr := fmt.Sprintf("%s:%d", address.IP, port.Port)
							hs[addr] = true

							//logger.Info(addr)
						}
						//logger.Info(address.IP)
					}
				}
			}
			if reflect.DeepEqual(my_hystrixs, hs) == false {
				logger.Info("ep changed", "old hystrixs:", my_hystrixs, "new hystrixs:", hs)

				my_hystrixs = hs

				pod := corev1.Pod{}
				// get monitor
				monitor_namespace_name := types.NamespacedName{Namespace: req.Namespace, Name: turbine_name}
				turbineMonitor := monitorwangjldevv1beta1.Turbine{}
				if err := r.Get(ctx, monitor_namespace_name, &turbineMonitor); err != nil {
					logger.Error(err, "cannot find", "turbineMonitor", monitor_namespace_name)
					return ctrl.Result{}, nil
				}
				if tmp_pod, err := constructTurbinePod(&turbineMonitor, r.Scheme); err != nil {
					logger.Error(err, "construct TurbinePod")
					return ctrl.Result{}, nil
				} else {
					pod = *tmp_pod
				}

				turbine_streams_str := strings.Builder{}
				for stream, _ := range my_hystrixs {
					turbine_streams_str.WriteString(stream)
					turbine_streams_str.WriteString(",")
				}
				turbine_streams_str.Len()
				pod.Annotations["turbine_streams"] = turbine_streams_str.String()[0 : turbine_streams_str.Len()-1]

				hystrix_streams_str := strings.Builder{}
				for stream, _ := range my_hystrixs {
					hystrix_streams_str.WriteString("http://")
					hystrix_streams_str.WriteString(stream)
					hystrix_streams_str.WriteString("/hystrix.stream ")
				}
				pod.Spec.Containers[0].Env[0].Value = hystrix_streams_str.String()[0 : hystrix_streams_str.Len()-1]

				if pod_exists {
					if old_pod.DeletionTimestamp != nil {
						logger.Info("in deleting", "pod", old_pod)
						return scheduledResult, nil
					}
					logger.Info("delete", "pod", old_pod)
					if err := r.Delete(ctx, &old_pod, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
						logger.Error(err, "delete", "pod", pod)
					}
					return scheduledResult, nil
				}
				logger.Info("create", "pod", pod)
				if err := r.Create(ctx, &pod); err != nil {
					logger.Error(err, "create", "pod", pod)
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *TurbineServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		//For(&monitorwangjldevv1beta1.TurbineServer{}).
		For(&corev1.Endpoints{}).
		Complete(r)
}
