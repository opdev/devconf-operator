package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var podSecContext = corev1.PodSecurityContext{
	RunAsNonRoot: &[]bool{true}[0],
	SeccompProfile: &corev1.SeccompProfile{
		Type: corev1.SeccompProfileTypeRuntimeDefault,
	},
}

var secContext = &corev1.SecurityContext{
	// WARNING: Ensure that the image used defines an UserID in the Dockerfile
	// otherwise the Pod will not run and will fail with `container has runAsNonRoot and image has non-numeric user`.
	// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
	// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the `RunAsNonRoot` and
	// RunAsUser fields empty.
	RunAsNonRoot:             &[]bool{true}[0],
	AllowPrivilegeEscalation: &[]bool{false}[0],
	Capabilities: &corev1.Capabilities{
		Drop: []corev1.Capability{
			"ALL",
		},
	},
}

var databaseImage = "image-registry.openshift-image-registry.svc:5000/openshift/mysql@sha256:8e9a6595ac9aec17c62933d3b5ecc78df8174a6c2ff74c7f602235b9aef0a340"

func MysqlDeploymentForRecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	if recipe.Spec.Database.PodSecurityContext != nil {
		podSecContext = *recipe.Spec.Database.PodSecurityContext
	}
	if recipe.Spec.Database.SecurityContext != nil {
		secContext = recipe.Spec.Database.SecurityContext
	}
	if recipe.Spec.Database.Image != "" {
		databaseImage = recipe.Spec.Database.Image
	}
	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name + "-mysql",
			Namespace: recipe.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": recipe.Name + "-mysql",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": recipe.Name + "-mysql",
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &podSecContext,
					Containers: []corev1.Container{{
						Image: databaseImage,
						Name:  "mysql",
						Args: []string{
							"--ignore-db-dir=lost+found",
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 3306,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name: "MYSQL_DATABASE",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: recipe.Name + "-mysql-config",
										},
										Key: "MYSQL_DATABASE",
									},
								},
							}, {
								Name: "MYSQL_USER",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: recipe.Name + "-mysql-config",
										},
										Key: "MYSQL_USER",
									},
								},
							}, {
								Name: "MYSQL_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: recipe.Name + "-mysql",
										},
										Key: "MYSQL_PASSWORD",
									},
								},
							}, {
								Name: "MYSQL_ROOT_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: recipe.Name + "-mysql",
										},
										Key: "MYSQL_ROOT_PASSWORD",
									},
								},
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "mysql-persistent-storage",
								MountPath: "/var/lib/mysql",
							},
							{
								Name:      "mysql-initdb",
								MountPath: "/docker-entrypoint-initdb.d",
							},
						},
						SecurityContext: secContext,
					}},
					Volumes: []corev1.Volume{
						{
							Name: "mysql-persistent-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: recipe.Name + "-mysql",
								},
							},
						},
						{Name: "mysql-initdb",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: recipe.Name + "-mysql-initdb-config",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	// Set the ownerRef for the Deployment
	if err := ctrl.SetControllerReference(recipe, dep, scheme); err != nil {
		return nil, err
	}
	return dep, nil
}
