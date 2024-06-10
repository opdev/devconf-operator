/*
Copyright 2024.

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

// KarbanatekSpec defines the desired state of Karbanatek
type KarbanatekSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Version is the version of the recipe app image to run
	Version string `json:"version,omitempty"`

	// Replicas is the number of replicas to run
	Count int32 `json:"count,omitempty"`
}

// KarbanatekStatus defines the observed state of Karbanatek
type KarbanatekStatus struct {
	MySQLStatus     string `json:"mysqlStatus,omitempty"`
	RecipeAppStatus string `json:"recipeAppStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Karbanatek is the Schema for the karbanateks API
type Karbanatek struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KarbanatekSpec   `json:"spec,omitempty"`
	Status KarbanatekStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KarbanatekList contains a list of Karbanatek
type KarbanatekList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Karbanatek `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Karbanatek{}, &KarbanatekList{})
}
