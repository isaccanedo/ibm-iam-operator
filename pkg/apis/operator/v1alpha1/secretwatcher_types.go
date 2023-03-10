//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretWatcherSpec defines the desired state of SecretWatcher
type SecretWatcherSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	OperatorVersion string                       `json:"operatorVersion,omitempty"`
	Replicas        int32                        `json:"replicas"`
	ImageRegistry   string                       `json:"imageRegistry,omitempty"`
	ImageTagPostfix string                       `json:"imageTagPostfix,omitempty"`
	Resources       *corev1.ResourceRequirements `json:"resources,omitempty"`
	Config          SecretWatcherConfigSpec      `json:"config,omitempty"`
}

type SecretWatcherConfigSpec struct {
	ExcludeOperand bool `json:"excludeOperand,omitempty"`
}

// SecretWatcherStatus defines the observed state of SecretWatcher
type SecretWatcherStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecretWatcher is the Schema for the secretwatchers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=secretwatchers,scope=Namespaced
type SecretWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretWatcherSpec   `json:"spec,omitempty"`
	Status SecretWatcherStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecretWatcherList contains a list of SecretWatcher
type SecretWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretWatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretWatcher{}, &SecretWatcherList{})
}
