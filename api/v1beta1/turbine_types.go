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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Hystrix struct {
	corev1.ServiceSpec `json:",inline"`
	Name               string `json:"name,omitempty"`
}

// TurbineSpec defines the desired state of Turbine
type TurbineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Turbine. Edit Turbine_types.go to remove/update
	//Foo      string               `json:"foo,omitempty"`
	//Hystrixs []corev1.ServiceSpec `json:"hystrixs"`
	//Hystrixs []corev1.Service `json:"hystrixs"`
	Hystrixs []Hystrix `json:"hystrixs"`
}

// TurbineStatus defines the observed state of Turbine
type TurbineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Turbine is the Schema for the turbines API
type Turbine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TurbineSpec   `json:"spec,omitempty"`
	Status TurbineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TurbineList contains a list of Turbine
type TurbineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Turbine `json:"items"`
}

//var HystrixList corev1.ServiceList

var MonitorName string

func init() {
	SchemeBuilder.Register(&Turbine{}, &TurbineList{})
}
