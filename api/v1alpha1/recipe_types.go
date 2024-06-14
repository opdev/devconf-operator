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
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RecipeSpec defines the desired state of Recipe
type RecipeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Version is the version of the recipe app image to run
	Version string `json:"version,omitempty"`

	// Replicas is the number of replicas to run
	Replicas int32 `json:"replicas,omitempty"`

	// Hpa specifies the pod autoscaling configuration to use
	// for the workload.
	// +optional
	Hpa *HpaSpec `json:"hpa,omitempty"`
}

type HpaSpec struct {
	// MinReplicas sets a lower bound to the autoscaling feature.  Set this if your are using autoscaling. It must be at least 1
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// MaxReplicas sets an upper bound to the autoscaling feature. If MaxReplicas is set autoscaling is enabled.
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
	// +optional
	Behavior *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
	// +optional
	// TargetMemoryUtilization sets the target average memory utilization across all replicas
	TargetMemoryUtilization *int32 `json:"targetMemoryUtilization,omitempty"`
}
// RecipeStatus defines the observed state of Recipe
type RecipeStatus struct {
	MySQLStatus     string `json:"mysqlStatus,omitempty"`
	RecipeAppStatus string `json:"recipeAppStatus,omitempty"`
	RecipeAppHpa    string `json:"recipeAppHpa,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Recipe is the Schema for the recipes API
type Recipe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecipeSpec   `json:"spec,omitempty"`
	Status RecipeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RecipeList contains a list of Recipe
type RecipeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Recipe `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Recipe{}, &RecipeList{})
}
