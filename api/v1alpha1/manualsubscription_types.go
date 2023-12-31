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

// ManualSubscriptionSpec defines the desired state of ManualSubscription
type ManualSubscriptionSpec struct {
	Channel         string `json:"channel,omitempty"`
	Name            string `json:"name,omitempty"`
	Source          string `json:"source,omitempty"`
	SourceNamespace string `json:"sourceNamespace,omitempty"`
	StartingCSV     string `json:"startingCSV,omitempty"`
}

// ManualSubscriptionStatus defines the observed state of ManualSubscription
type ManualSubscriptionStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	SubscriptionRef    string             `json:"subscriptionRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ManualSubscription is the Schema for the manualsubscriptions API
type ManualSubscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManualSubscriptionSpec   `json:"spec,omitempty"`
	Status ManualSubscriptionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManualSubscriptionList contains a list of ManualSubscription
type ManualSubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManualSubscription `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManualSubscription{}, &ManualSubscriptionList{})
}
