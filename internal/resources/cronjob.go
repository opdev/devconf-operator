// apiVersion: batch/v1beta1
// kind: CronJob
// metadata:
//   name: mariadb-backup-cronjob
// spec:
//   schedule: "0 0 * * *"
//   jobTemplate:
//     spec:
//       template:
//         spec:
//           containers:
//             - name: mariadb-backup
//               image: mysql:latest
//               command:
//                 - "/bin/sh"
//                 - "-c"
//                 - "mysqldump -h mariadb-service -u $(cat /secrets/username) -p$(cat /secrets/password) --all-databases > /backup/backup-$(date +'%Y%m%d%H%M%S').sql"
//               volumeMounts:
//                 - name: backup-volume
//                   mountPath: /backup
//                 - name: mariadb-secrets
//                   mountPath: /secrets
//                   readOnly: true
//           restartPolicy: OnFailure
//           volumes:
//             - name: backup-volume
//               persistentVolumeClaim:
//                 claimName: mariadb-pvc
//             - name: mariadb-secrets
//               secret:
//                 secretName: mariadb-credentials

package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	// "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"strconv"
)

// CronJobForBackups creates a CronJob for MySQL backup ups and sets the owner reference
func CronJobForBackups(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*batchv1.CronJob, error) {
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name + "-mysql-backup",
			Namespace: recipe.Namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "0/" + strconv.Itoa(int(recipe.Spec.BackupPolicy.Interval)) + " * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							// SecurityContext: &corev1.PodSecurityContext{
							// 	RunAsNonRoot: &[]bool{true}[0],
							// 	SeccompProfile: &corev1.SeccompProfile{
							// 		Type: corev1.SeccompProfileTypeRuntimeDefault,
							// 	},
							// },
							Containers: []corev1.Container{{
								Image:           "image-registry.openshift-image-registry.svc:5000/openshift/mysql@sha256:8e9a6595ac9aec17c62933d3b5ecc78df8174a6c2ff74c7f602235b9aef0a340",
								Name:            "mysql-backup",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Command: []string{
									"/bin/sh",
									"-c",
									"mysqldump -u $recipeuser -p$MYSQL_PASSWORD $MYSQL_DATABASE > /backup/backup-$(date +'%Y%m%d%H%M%S').sql",
								},
								Env: []corev1.EnvVar{
									{
										Name:  "MYSQL_ROOT_PASSWORD",
										Value: "rootpassword",
									},
									{
										Name:  "MYSQL_DATABASE",
										Value: "recipes",
									},
									{
										Name:  "MYSQL_PASSWORD",
										Value: "recipepassword",
									},
									{
										Name:  "MYSQL_USER",
										Value: "recipeuser",
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "mysql-persistent-storage",
										MountPath: "/var/lib/mysql",
									},
									{
										Name:      "backup-storage",
										MountPath: "/backup",
									},
								},
								// SecurityContext: &corev1.SecurityContext{
								// 	// WARNING: Ensure that the image used defines an UserID in the Dockerfile
								// 	// otherwise the Pod will not run and will fail with `container has runAsNonRoot and image has non-numeric user`.
								// 	// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
								// 	// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the `RunAsNonRoot` and
								// 	// RunAsUser fields empty.
								// 	RunAsNonRoot:             &[]bool{true}[0],
								// 	AllowPrivilegeEscalation: &[]bool{false}[0],
								// 	Capabilities: &corev1.Capabilities{
								// 		Drop: []corev1.Capability{
								// 			"ALL",
								// 		},
								// 	},
								// },
							}},
							Volumes: []corev1.Volume{
								{
									Name: "mysql-persistent-storage",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: "mysql",
										},
									},
								}, {
									Name: "backup-storage",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: recipe.Spec.BackupPolicy.BackupPVC,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(recipe, cj, scheme); err != nil {
		return nil, err
	}

	return cj, nil
}
