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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	resources "github.com/opdev/devconf-operator/internal/resources"
)

// recipeReconciler reconciles a recipe object
type RecipeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=*
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses;servicemonitors;prometheusrule,verbs=*
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps;endpoints;events;persistentvolumeclaims;pods;namespaces;secrets;serviceaccounts;services;services/finalizers,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the recipe object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *RecipeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := log.FromContext(ctx)

	// get an instance of the recipe object
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
	// Define a new ConfigMap object for initdbconfigmap mysql database
	mysqlInitDBConfigMap, err := resources.MySQLInitDBConfigMapForrecipe(recipe, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Check if the InitDB ConfigMap already exists
	foundMysqlInitDBConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, client.ObjectKey{Name: mysqlInitDBConfigMap.Name, Namespace: mysqlInitDBConfigMap.Namespace}, foundMysqlInitDBConfigMap)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new ConfigMap for mysql database initialization")
		err = r.Create(ctx, mysqlInitDBConfigMap)
		if err != nil {
			log.Error(err, "Failed to create new ConfigMap for mysql database initialization", "ConfigMap.Namespace", mysqlInitDBConfigMap.Namespace, "ConfigMap.Name", mysqlInitDBConfigMap.Name)
			return ctrl.Result{}, err
		}
		// ConfigMap created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ConfigMap for mysql database initialization")
		return ctrl.Result{}, err
	}

	// Define a new ConfigMap object for mysql database
	mysqlConfigMap, err := resources.MySQLConfigMapForrecipe(recipe, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Check if the ConfigMap already exists
	foundMySQLConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, client.ObjectKey{Name: mysqlConfigMap.Name, Namespace: mysqlConfigMap.Namespace}, foundMySQLConfigMap)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new MySQL ConfigMap", "ConfigMap.Namespace", mysqlConfigMap.Namespace, "ConfigMap.Name", mysqlConfigMap.Name)
		err = r.Create(ctx, mysqlConfigMap)
		if err != nil {
			log.Error(err, "Failed to create new MySQL ConfigMap", "ConfigMap.Namespace", mysqlConfigMap.Namespace, "ConfigMap.Name", mysqlConfigMap.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get MySQL ConfigMap")
		return ctrl.Result{}, err
	}

	// Define a new service object for recipe application
	service, err := resources.RecipeServiceForrecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define new service resource for recipe application")
		return ctrl.Result{}, err
	}
	// Check if the service already exists
	foundService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new service for recipe application")
		err = r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create new service for recipe application", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get service")
		return ctrl.Result{}, err
	}

	// Define a new service object for mysql database
	service, err = resources.MySQLServiceForrecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define new service resource for mysql database")
		return ctrl.Result{}, err
	}
	// Check if the service already exists
	foundService = &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new service resource for mysql database")
		err = r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create new service for mysql database", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get service for mysql database")
		return ctrl.Result{}, err
	}

	// Define a new persistent volume claim object
	pvc, err := resources.PersistentVolumeClaimForrecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define PVC for recipe")
		return ctrl.Result{}, err
	}
	// Check if the PVC already exists
	foundPVC := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, client.ObjectKey{Name: pvc.Name, Namespace: pvc.Namespace}, foundPVC)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new PVC")
		err = r.Create(ctx, pvc)
		if err != nil {
			log.Error(err, "Failed to create new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
			return ctrl.Result{}, err
		}
		// PVC created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PVC")
		return ctrl.Result{}, err
	}

	// Define a new mysql database Deployment object
	dep, err := resources.MysqlDeploymentForrecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define new mysql deployment resource for recipe")
		return ctrl.Result{}, err
	}

	// Check if the Mysql database Deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// Update status for MySQL Deployment
		recipe.Status.MySQLStatus = "Creating..."
		log.Info("Creating a new mysql database deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new mysql database deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)

			// Update status for MySQL Deployment
			recipe.Status.MySQLStatus = "Created"
			err = r.Status().Update(ctx, recipe)
			if err != nil {
				log.Error(err, "Failed to update recipe status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get mysql database deployment")
		// Update status for MySQL Deployment
		recipe.Status.MySQLStatus = "Failed"
		return ctrl.Result{}, err
	}
	// mysql database Deployment already exists - don't requeue
	log.Info("Skip reconcile: mysql database deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

	// Define a new recipe app deployment object
	dep, err = resources.DeploymentForRecipeApp(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define new Deployment resource for recipe")
		return ctrl.Result{}, err
	}

	// Check if the Deployment already exists
	found = &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {

		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)

			// Update status for Recipe App Deployment
			recipe.Status.RecipeAppStatus = "Created"
			err = r.Status().Update(ctx, recipe)
			if err != nil {
				log.Error(err, "Failed to update recipe status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	} else if *found.Spec.Replicas != recipe.Spec.Count {
		// Update the Recipe deployment if the number of replicas does not match the desired state
		log.Info("Updating Recipe Deployment replicas", "Current", *found.Spec.Replicas, "Desired", recipe.Spec.Count)
		found.Spec.Replicas = &recipe.Spec.Count
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Recipe Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
	}

	// Update status for MySQL Deployment
	recipe.Status.MySQLStatus = "Created"
	// Update status for Recipe App Deployment
	recipe.Status.RecipeAppStatus = "Created"
	err = r.Status().Update(ctx, recipe)
	if err != nil {
		log.Error(err, "Failed to update recipe status")
		return ctrl.Result{}, err
	}

	// If the Deployment already exists and the size is the same, then do nothing
	log.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecipeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devconfczv1alpha1.Recipe{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
