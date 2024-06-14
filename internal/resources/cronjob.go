package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MySQLConfigMapForrecipe creates a ConfigMap for MySQL configuration
func CronJobForMySqlBackup(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*batchv1.CronJob, error) {
	var cronJob *batchv1.CronJob
	if recipe.Spec.Database.BackupPolicy.Schedule != "" {
		cronJob = &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysql-job",
				Namespace: recipe.Namespace,
			},
			Spec: batchv1.CronJobSpec{
				Schedule: recipe.Spec.Database.BackupPolicy.Schedule,
				TimeZone: &recipe.Spec.Database.BackupPolicy.Tmz,
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:           "fradelg/mysql-cron-backup",
									Name:            "job-mysql",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Env: []corev1.EnvVar{
										{
											Name: "MYSQL_HOST",
											ValueFrom: &corev1.EnvVarSource{
												ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "mysql." + recipe.Namespace + ".svc",
													},
													Key: "MYSQL_HOST",
												},
											},
										},
										{
											Name: "MYSQL_PORT",
											ValueFrom: &corev1.EnvVarSource{
												ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "mysql-config",
													},
													Key: "MYSQL_PORT",
												},
											},
										},
										{
											Name: "MYSQL_NAME",
											ValueFrom: &corev1.EnvVarSource{
												ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "mysql-config",
													},
													Key: "MYSQL_DATABASE",
												},
											},
										},
										{
											Name:  "MYSQL_USER",
											Value: "recipeuser",
										},
										{
											Name: "MYSQL_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "mysql-config",
													},
													Key: "MYSQL_PASSWORD",
												},
											},
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "mysql-backup",
											MountPath: "/backup",
										},
									},
								}},
								Volumes: []corev1.Volume{
									{
										Name: "mysql-backup",
										VolumeSource: corev1.VolumeSource{
											PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
												ClaimName: "mysql-backup",
											},
										},
									},
								},
								RestartPolicy: "OnFailure",
							},
						},
					},
				},
			},
		}
	}
	if err := ctrl.SetControllerReference(recipe, cronJob, scheme); err != nil {
		return nil, err
	}

	return cronJob, nil
}
