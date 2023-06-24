/*
Copyright 2023.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CrdRefSpec defines the desired state of CrdRef
type CrdRefSpec struct {
	// crd url or yaml
	Crd string `json:"crd,omitempty"`
}

// CrdRefStatus defines the observed state of CrdRef
type CrdRefStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	CrdRef             string             `json:"crdRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CrdRef is the Schema for the crdrefs API
type CrdRef struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CrdRefSpec   `json:"spec,omitempty"`
	Status CrdRefStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CrdRefList contains a list of CrdRef
type CrdRefList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CrdRef `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CrdRef{}, &CrdRefList{})
}
