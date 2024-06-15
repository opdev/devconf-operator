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

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	resources "github.com/opdev/devconf-operator/internal/resources"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

// RecipeReconciler reconciles a Recipe object
type RecipeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devconfcz.opdev.com,resources=recipes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=*
//+kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=*
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses;servicemonitors;prometheusrule,verbs=*
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps;endpoints;events;persistentvolumeclaims;pods;namespaces;secrets;serviceaccounts;services;services/finalizers,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Recipe object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *RecipeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := log.FromContext(ctx)
	imagename := "quay.io/opdev/recipe_app"

	// get an instance of the Recipe object
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
	mysqlInitDBConfigMap, err := resources.MySQLInitDBConfigMapForRecipe(recipe, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Check if the InitDB ConfigMap already exists
	err = r.Get(ctx, client.ObjectKey{Name: mysqlInitDBConfigMap.Name, Namespace: mysqlInitDBConfigMap.Namespace}, &corev1.ConfigMap{})
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
	mysqlConfigMap, err := resources.MySQLConfigMapForRecipe(recipe, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Check if the ConfigMap already exists
	err = r.Get(ctx, client.ObjectKey{Name: mysqlConfigMap.Name, Namespace: mysqlConfigMap.Namespace}, &corev1.ConfigMap{})
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

	// Define a new Secret object for mysql database
	mysqlSecret, err := resources.MySQLSecretForRecipe(recipe, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Check if the Secret already exists
	err = r.Get(ctx, client.ObjectKey{Name: mysqlSecret.Name, Namespace: mysqlSecret.Namespace}, &corev1.Secret{})
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new Secret for mysql")
		err = r.Create(ctx, mysqlSecret)
		if err != nil {
			log.Error(err, "Failed to create new Secret for mysql database initialization", "Secret.Namespace", mysqlSecret.Namespace, "Secret.Name", mysqlSecret.Name)
			return ctrl.Result{}, err
		}
		// Secret created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Secret for mysql database initialization")
		return ctrl.Result{}, err
	}

	// Define a new service object for recipe application
	service, err := resources.RecipeServiceForRecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define new service resource for recipe application")
		return ctrl.Result{}, err
	}
	// Check if the service already exists
	err = r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, &corev1.Service{})
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
	service, err = resources.MySQLServiceForRecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define new service resource for mysql database")
		return ctrl.Result{}, err
	}
	// Check if the service already exists
	err = r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, &corev1.Service{})
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
	pvc, err := resources.PersistentVolumeClaimForRecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define PVC for recipe")
		return ctrl.Result{}, err
	}
	// Check if the PVC already exists
	err = r.Get(ctx, client.ObjectKey{Name: pvc.Name, Namespace: pvc.Namespace}, &corev1.PersistentVolumeClaim{})
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
	dep, err := resources.MysqlDeploymentForRecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define new mysql deployment resource for recipe")
		return ctrl.Result{}, err
	}

	// Check if the Mysql database Deployment already exists
	err = r.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, &appsv1.Deployment{})
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

	// Define a new recipe app deployment object
	dep, err = resources.DeploymentForRecipe(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define new Deployment resource for recipe")
		return ctrl.Result{}, err
	}

	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
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
	} else if *found.Spec.Replicas != recipe.Spec.Replicas {
		// Update the Recipe deployment if the number of replicas does not match the desired state
		log.Info("Updating Recipe Deployment replicas", "Current", *found.Spec.Replicas, "Desired", recipe.Spec.Replicas)
		found.Spec.Replicas = &recipe.Spec.Replicas
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Recipe Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
	}

	// If the Deployment already exists and the size is the same, then do nothing
	log.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)


	// Level 2: Update Operand (Recipe App)
	log.Info("Reconciling Recipe App version")
	found = &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{Name: recipe.Name, Namespace: recipe.Namespace}, found)

	if err != nil {
		log.Error(err, "Failed to get Recipe App Deployment")
		return ctrl.Result{}, err
	}
	desiredImage := fmt.Sprintf("%s:%s", imagename, recipe.Spec.Version)
	currentImage := found.Spec.Template.Spec.Containers[0].Image

	if currentImage != desiredImage {
		// Level 4 Increment the upgrades metric
		upgrades.Inc()

		found.Spec.Template.Spec.Containers[0].Image = desiredImage
		err = r.Update(ctx, found)
		if err != nil {
			// Level 4 Increment the upgradesFailures metric
			upgradesFailures.Inc()
			log.Error(err, "Failed to update Recipe App version")
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

	hpa, err := resources.AutoScaler(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to create a HorizontalPodAutoscaler resource for recipe")
		return ctrl.Result{}, err
	}

	foundHpa := &autoscalingv2.HorizontalPodAutoscaler{}
	err = r.Get(ctx, client.ObjectKey{Name: hpa.Name, Namespace: hpa.Namespace}, foundHpa)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new HorizontalPodAutoScaler", "HorizontalPodAutoScaler.Namespace", hpa.Namespace, "HorizontalPodAutoScaler.Name", hpa.Name)
		err = r.Create(ctx, hpa)
		if err != nil {
			log.Error(err, "Failed to create new HorizontalPodAutoScaler", "HorizontalPodAutoScaler.Namespace", hpa.Namespace, "HorizontalPodAutoScaler.Name", hpa.Name)

			// Update status for Recipe App HorizontalPodAutoScaler
			recipe.Status.RecipeAppHpa = "HPA Created"
			err = r.Status().Update(ctx, recipe)
			if err != nil {
				log.Error(err, "Failed to update recipe hpa status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}
		// HorizontalPodAutoScaler created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to filter HorizontalPodAutoScaler")
		return ctrl.Result{}, err
	}

	pvcCronJob, err := resources.PersistentVolumeClaimForBackup(recipe, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to define PVC-CronJob for recipe")
		return ctrl.Result{}, err
	}
	// Check if the pvcCronJob already exists
	err = r.Get(ctx, client.ObjectKey{Name: pvcCronJob.Name, Namespace: pvcCronJob.Namespace}, &corev1.PersistentVolumeClaim{})
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new pvcCronJob")
		err = r.Create(ctx, pvcCronJob)
		if err != nil {
			log.Error(err, "Failed to create new pvcCronJob", "pvcCronJob.Namespace", pvcCronJob.Namespace, "pvcCronJob.Name", pvcCronJob.Name)
			return ctrl.Result{}, err
		}
		// pvcCronJob created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get pvcCronJob")
		return ctrl.Result{}, err
	}

	if recipe.Spec.Database.BackupPolicy.Schedule != "" {
		cronJob, err := resources.CronJobForMySqlBackup(recipe, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to create a CronJob Backup resource for recipe")
			return ctrl.Result{}, err
		}

		foundCronJob := &batchv1.CronJob{}
		err = r.Get(ctx, client.ObjectKey{Name: cronJob.Name, Namespace: cronJob.Namespace}, foundCronJob)
		if err != nil && apierrors.IsNotFound(err) {
			log.Info("Creating a new CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
			err = r.Create(ctx, cronJob)
			if err != nil {
				log.Error(err, "Failed to create new CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
				return ctrl.Result{}, err
			}
			// CronJob created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to filter CronJob")
			return ctrl.Result{}, err
		}
	}

	if recipe.Spec.Database.InitRestore {
		job, err := resources.JobForMySqlRestore(recipe, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to define Restore Job for recipe")
			return ctrl.Result{}, err
		}
		// Check if the job already exists
		foundJob := &batchv1.Job{}
		err = r.Get(ctx, client.ObjectKey{Name: job.Name, Namespace: job.Namespace}, foundJob)
		if err != nil && apierrors.IsNotFound(err) {
			log.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			err = r.Create(ctx, job)
			if err != nil {
				log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
				return ctrl.Result{}, err
			}
			// Job created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to filter Job")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecipeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devconfczv1alpha1.Recipe{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&batchv1.CronJob{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
