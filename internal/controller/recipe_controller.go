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

package controller

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
)

const recipeFinalizer = "devconfcz.opdev.com/finalizer"

// Definitions to manage status conditions
const (
	// typeAvailableRecipe represents the status of the StatefulSet reconciliation
	typeAvailableRecipe = "Available"
	// typeDegradedRecipe represents the status used when the custom resource is deleted and the finalizer operations are must to occur.
	typeDegradedRecipe = "Degraded"
)

// RecipeReconciler reconciles a Recipe object
type RecipeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// The following markers are used to generate the rules permissions (RBAC) on config/rbac using controller-gen
// when the command <make manifests> is executed.
// To know more about markers see: https://book.kubebuilder.io/reference/markers.html

//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It is essential for the controller's reconciliation loop to be idempotent. By following the Operator
// pattern you will create Controllers which provide a reconcile function
// responsible for synchronizing resources until the desired state is reached on the cluster.
// Breaking this recommendation goes against the design principles of controller-runtime.
// and may lead to unforeseen consequences such as resources becoming stuck and requiring manual intervention.
// For further info:
// - About Operator Pattern: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
// - About Controllers: https://kubernetes.io/docs/concepts/architecture/controller/
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *RecipeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Recipe instance
	// The purpose is check if the Custom Resource for the Kind Recipe
	// is applied on the cluster if not we return nil to stop the reconciliation
	recipe := &devconfczv1alpha1.Recipe{}
	err := r.Get(ctx, req.NamespacedName, recipe)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("recipe resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get recipe")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status are available
	if recipe.Status.Conditions == nil || len(recipe.Status.Conditions) == 0 {
		meta.SetStatusCondition(&recipe.Status.Conditions, metav1.Condition{Type: typeAvailableRecipe, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, recipe); err != nil {
			log.Error(err, "Failed to update Recipe status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the recipe Custom Resource after update the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raise the issue "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, recipe); err != nil {
			log.Error(err, "Failed to re-fetch recipe")
			return ctrl.Result{}, err
		}
	}

	// Check if the Recipe instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isRecipeMarkedToBeDeleted := recipe.GetDeletionTimestamp() != nil
	if isRecipeMarkedToBeDeleted {
		log.Info("Ignoring Recipe being deleted")
		return ctrl.Result{}, nil
	}

	// Ensure child resources are present

	// Check if the Service already exists, if not create a new one
	err = r.Get(ctx, types.NamespacedName{Name: recipe.Name, Namespace: recipe.Namespace}, &corev1.Service{})
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new service
		svc, err := r.serviceForRecipe(recipe)
		if err != nil {
			log.Error(err, "Failed to define new Service resource for Recipe")

			// The following implementation will update the status
			meta.SetStatusCondition(&recipe.Status.Conditions, metav1.Condition{Type: typeAvailableRecipe,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create Service for the custom resource (%s): (%s)", recipe.Name, err)})

			if err := r.Status().Update(ctx, recipe); err != nil {
				log.Error(err, "Failed to update Recipe status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		log.Info("Creating a new Service",
			"Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		if err = r.Create(ctx, svc); err != nil {
			log.Error(err, "Failed to create new Service",
				"Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}

		// Service created successfully
		// We will requeue the reconciliation so that we can ensure the state
		// and move forward for the next operations
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		// Let's return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// Check if the PVC already exists, if not create a new one
	err = r.Get(ctx, types.NamespacedName{Name: recipe.Name, Namespace: recipe.Namespace}, &corev1.PersistentVolumeClaim{})
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new pvc
		pvc, err := r.pvcForRecipe(recipe)
		if err != nil {
			log.Error(err, "Failed to define new PVC resource for Recipe")

			// The following implementation will update the status
			meta.SetStatusCondition(&recipe.Status.Conditions, metav1.Condition{Type: typeAvailableRecipe,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create PVC for the custom resource (%s): (%s)", recipe.Name, err)})

			if err := r.Status().Update(ctx, recipe); err != nil {
				log.Error(err, "Failed to update Recipe status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		log.Info("Creating a new PVC",
			"PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
		if err = r.Create(ctx, pvc); err != nil {
			log.Error(err, "Failed to create new PVC",
				"PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
			return ctrl.Result{}, err
		}

		// PVC created successfully
		// We will requeue the reconciliation so that we can ensure the state
		// and move forward for the next operations
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PVC")
		// Let's return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// Check if the statefulset already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: recipe.Name, Namespace: recipe.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new statefulset
		sts, err := r.statefulsetForRecipe(recipe)
		if err != nil {
			log.Error(err, "Failed to define new StatefulSet resource for Recipe")

			// The following implementation will update the status
			meta.SetStatusCondition(&recipe.Status.Conditions, metav1.Condition{Type: typeAvailableRecipe,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create StatefulSet for the custom resource (%s): (%s)", recipe.Name, err)})

			if err := r.Status().Update(ctx, recipe); err != nil {
				log.Error(err, "Failed to update Recipe status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		log.Info("Creating a new StatefulSet",
			"StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		if err = r.Create(ctx, sts); err != nil {
			log.Error(err, "Failed to create new StatefulSet",
				"StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
			return ctrl.Result{}, err
		}

		// StatefulSet created successfully
		// We will requeue the reconciliation so that we can ensure the state
		// and move forward for the next operations
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get StatefulSet")
		// Let's return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// The CRD API is defining that the Recipe type, have a RecipeSpec.Size field
	// to set the quantity of StatefulSet instances is the desired state on the cluster.
	// Therefore, the following code will ensure the StatefulSet size is the same as defined
	// via the Size spec of the Custom Resource which we are reconciling.
	size := recipe.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		if err = r.Update(ctx, found); err != nil {
			log.Error(err, "Failed to update StatefulSet",
				"StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)

			// Re-fetch the recipe Custom Resource before update the status
			// so that we have the latest state of the resource on the cluster and we will avoid
			// raise the issue "the object has been modified, please apply
			// your changes to the latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, recipe); err != nil {
				log.Error(err, "Failed to re-fetch recipe")
				return ctrl.Result{}, err
			}

			// The following implementation will update the status
			meta.SetStatusCondition(&recipe.Status.Conditions, metav1.Condition{Type: typeAvailableRecipe,
				Status: metav1.ConditionFalse, Reason: "Resizing",
				Message: fmt.Sprintf("Failed to update the size for the custom resource (%s): (%s)", recipe.Name, err)})

			if err := r.Status().Update(ctx, recipe); err != nil {
				log.Error(err, "Failed to update Recipe status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		// Now, that we update the size we want to requeue the reconciliation
		// so that we can ensure that we have the latest state of the resource before
		// update. Also, it will help ensure the desired state on the cluster
		return ctrl.Result{Requeue: true}, nil
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&recipe.Status.Conditions, metav1.Condition{Type: typeAvailableRecipe,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("StatefulSet for custom resource (%s) with %d replicas created successfully", recipe.Name, size)})

	if err := r.Status().Update(ctx, recipe); err != nil {
		log.Error(err, "Failed to update Recipe status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// statefulsetForRecipe returns a Recipe StatefulSet object
func (r *RecipeReconciler) statefulsetForRecipe(
	recipe *devconfczv1alpha1.Recipe) (*appsv1.StatefulSet, error) {
	ls := labelsForRecipe(recipe.Name)
	replicas := recipe.Spec.Size

	// Get the Operand image
	image, err := imageForRecipe()
	if err != nil {
		return nil, err
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name,
			Namespace: recipe.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			ServiceName: recipe.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					// SecurityContext: &corev1.PodSecurityContext{
					// 	RunAsNonRoot: &[]bool{true}[0],
					// 	// IMPORTANT: seccomProfile was introduced with Kubernetes 1.19
					// 	// If you are looking for to produce solutions to be supported
					// 	// on lower versions you must remove this option.
					// 	SeccompProfile: &corev1.SeccompProfile{
					// 		Type: corev1.SeccompProfileTypeRuntimeDefault,
					// 	},
					// },
					Containers: []corev1.Container{{
						Image:           image,
						Name:            "recipe",
						ImagePullPolicy: corev1.PullIfNotPresent,
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						// SecurityContext: &corev1.SecurityContext{
						// 	RunAsNonRoot:             &[]bool{true}[0],
						// 	AllowPrivilegeEscalation: &[]bool{false}[0],
						// 	Capabilities: &corev1.Capabilities{
						// 		Drop: []corev1.Capability{
						// 			"ALL",
						// 		},
						// 	},
						// },
						Ports: []corev1.ContainerPort{{
							ContainerPort: recipe.Spec.ContainerPort,
							Name:          "recipe",
						}},
						Env: []corev1.EnvVar{{
							Name:  "DATABASE_PATH",
							Value: "/database/recipes.db",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "recipe-db-storage",
							MountPath: "/database",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "recipe-db-storage",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: recipe.Name,
							},
						},
					}},
				},
			},
		},
	}

	// Set the ownerRef for the StatefulSet
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(recipe, sts, r.Scheme); err != nil {
		return nil, err
	}
	return sts, nil
}

// serviceForRecipe returns a Recipe Service object
func (r *RecipeReconciler) serviceForRecipe(
	recipe *devconfczv1alpha1.Recipe) (*corev1.Service, error) {
	ls := labelsForRecipe(recipe.Name)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name,
			Namespace: recipe.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       recipe.Spec.ContainerPort,
				TargetPort: intstr.FromInt32(recipe.Spec.ContainerPort),
			}},
			Selector: ls,
		},
	}

	// Set the ownerRef for the Service
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(recipe, svc, r.Scheme); err != nil {
		return nil, err
	}
	return svc, nil
}

// pvcForRecipe returns a Recipe Service object
func (r *RecipeReconciler) pvcForRecipe(
	recipe *devconfczv1alpha1.Recipe) (*corev1.PersistentVolumeClaim, error) {

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name,
			Namespace: recipe.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": *resource.NewQuantity(1*1024*1024*1024, resource.BinarySI),
				},
			},
		},
	}

	// Set the ownerRef for the PVC
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(recipe, pvc, r.Scheme); err != nil {
		return nil, err
	}
	return pvc, nil
}

// labelsForRecipe returns the labels for selecting the resources
// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
func labelsForRecipe(name string) map[string]string {
	var imageTag string
	image, err := imageForRecipe()
	if err == nil {
		imageTag = strings.Split(image, ":")[1]
	}
	return map[string]string{"app.kubernetes.io/name": "Recipe",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/version":    imageTag,
		"app.kubernetes.io/part-of":    "devconf-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}

// imageForRecipe gets the Operand image which is managed by this controller
// from the RECIPE_IMAGE environment variable defined in the config/manager/manager.yaml
func imageForRecipe() (string, error) {
	var imageEnvVar = "RECIPE_IMAGE"
	image, found := os.LookupEnv(imageEnvVar)
	if !found {
		return "", fmt.Errorf("Unable to find %s environment variable with the image", imageEnvVar)
	}
	return image, nil
}

// SetupWithManager sets up the controller with the Manager.
// Note that the StatefulSet will be also watched in order to ensure its
// desirable state on the cluster
func (r *RecipeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devconfczv1alpha1.Recipe{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
