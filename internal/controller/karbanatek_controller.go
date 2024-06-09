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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	resources "github.com/opdev/devconf-operator/resources"
)

// KarbanatekReconciler reconciles a Karbanatek object
type KarbanatekReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=karbanateks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=karbanateks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=karbanateks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Karbanatek object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *KarbanatekReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := log.FromContext(ctx)

	// get an instance of the Karbanatek object
	karbanatek := &devconfczv1alpha1.Karbanatek{}
	err := r.Get(ctx, req.NamespacedName, karbanatek)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("karbanatek resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get karbanatek")
		return ctrl.Result{}, err
	}
	// Define a new Deployment object
	dep, err := resources.DeploymentForKarbanatek(karbanatek)
	if err != nil {
		log.Error(err, "Failed to define new Deployment resource for Karbanatek")
		return ctrl.Result{}, err
	}

	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {

		// Ensure the Deployment size is the same as the spec
		size := karbanatek.Spec.Count
		if *found.Spec.Replicas != size {
			found.Spec.Replicas = &size
			err = r.Update(ctx, found)
			if err != nil {
				log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
				return ctrl.Result{}, err
			}
			// Spec updated - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}
	// If the Deployment already exists and the size is the same, then do nothing
	log.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

	return ctrl.Result{}, nil
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *KarbanatekReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devconfczv1alpha1.Karbanatek{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
