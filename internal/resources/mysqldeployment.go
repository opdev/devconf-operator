package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func MysqlDeploymentForrecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*appsv1.Deployment, error) {

	replicas := recipe.Spec.Count

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-deployment",
			Namespace: recipe.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "mysql",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "mysql",
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{{
						Image:           "image-registry.openshift-image-registry.svc:5000/openshift/mysql@sha256:8e9a6595ac9aec17c62933d3b5ecc78df8174a6c2ff74c7f602235b9aef0a340",
						Name:            "mysql",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 3306,
							},
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
								Name:      "mysql-initdb",
								MountPath: "/docker-entrypoint-initdb.d",
							},
						},
						SecurityContext: &corev1.SecurityContext{
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
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "mysql-persistent-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "mysql",
								},
							},
						},
						{Name: "mysql-initdb",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "mysql-initdb-config",
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
