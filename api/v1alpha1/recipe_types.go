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
	corev1 "k8s.io/api/core/v1"
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

	// PodSecurityContext in case of Openshift
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// SecurityContext in case of Openshift
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

	// Resources to set for Level 3 and 5.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Hpa specifies the pod autoscaling configuration to use
	// for the workload.
	// +optional
	Hpa *HpaSpec `json:"hpa,omitempty"`

	// Database specifies the database configuration to use
	// for the workload.
	// +optional
	Database DatabaseSpec `json:"database,omitempty"`
}

type HpaSpec struct {
	// MinReplicas sets a lower bound to the autoscaling feature.  Set this if your are using autoscaling. It must be at least 1
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// MaxReplicas sets an upper bound to the autoscaling feature. If MaxReplicas is set autoscaling is enabled.
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
	// +optional
	// TargetMemoryUtilization sets the target average memory utilization across all replicas
	TargetMemoryUtilization *int32 `json:"targetMemoryUtilization,omitempty"`
}

type DatabaseSpec struct {
	// VolumeName which should be used at MySQL DB.
	// +optional
	VolumeName string `json:"volumeName,omitempty"`
	// Image set the image which should be used at MySQL DB.
	// +optional
	Image string `json:"image,omitempty"`
	// PodSecurityContext in case of Openshift
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// SecurityContext in case of Openshift
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// BackupPolicy
	// +optional
	BackupPolicy BackupPolicySpec `json:"backupPolicySpec,omitempty"`
}

type BackupPolicySpec struct {
	// Backup Schedule
	// +optional
	Schedule string `json:"schedule,omitempty"`
	// Backup Schedule
	// +optional
	Tmz string `json:"timezone,omitempty"`
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
